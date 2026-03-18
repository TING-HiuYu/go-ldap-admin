package logic

import (
	"archive/zip"
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/eryajf/go-ldap-admin/config"
	"github.com/eryajf/go-ldap-admin/logic/webhook"
	"github.com/eryajf/go-ldap-admin/model"
	"github.com/eryajf/go-ldap-admin/model/request"
	"github.com/eryajf/go-ldap-admin/model/response"
	"github.com/eryajf/go-ldap-admin/public/common"
	"github.com/eryajf/go-ldap-admin/public/tools"
	"github.com/eryajf/go-ldap-admin/service/ildap"
	"github.com/eryajf/go-ldap-admin/service/isql"

	"github.com/gin-gonic/gin"
	"github.com/thoas/go-funk"
	"golang.org/x/crypto/ssh"
)

type UserLogic struct{}

// Add 添加数据
func (l UserLogic) Add(c *gin.Context, req any) (data any, rspError any) {
	r, ok := req.(*request.UserAddReq)
	if !ok {
		return nil, ReqAssertErr
	}
	_ = c

	if isql.User.Exist(tools.H{"mobile": r.Mobile}) {
		return nil, tools.NewValidatorError(fmt.Errorf("手机号已存在,请勿重复添加"))
	}
	if isql.User.Exist(tools.H{"job_number": r.JobNumber}) {
		return nil, tools.NewValidatorError(fmt.Errorf("工号已存在,请勿重复添加"))
	}
	if isql.User.Exist(tools.H{"mail": r.Mail}) {
		return nil, tools.NewValidatorError(fmt.Errorf("邮箱已存在,请勿重复添加"))
	}

	// 密码通过RSA解密
	// 密码不为空就解密
	if r.Password != "" {
		decodeData, err := tools.RSADecrypt([]byte(r.Password), config.Conf.System.RSAPrivateBytes)
		if err != nil {
			return nil, tools.NewValidatorError(fmt.Errorf("密码解密失败"))
		}
		r.Password = string(decodeData)
		if len(r.Password) < 6 {
			return nil, tools.NewValidatorError(fmt.Errorf("密码长度至少为6位"))
		}
	} else {
		r.Password = config.Conf.Ldap.UserInitPassword
	}

	// 当前登陆用户角色排序最小值（最高等级角色）以及当前登陆的用户
	currentRoleSortMin, ctxUser, err := isql.User.GetCurrentUserMinRoleSort(c)
	if err != nil {
		return nil, tools.NewValidatorError(fmt.Errorf("获取当前登陆用户角色排序最小值失败"))
	}

	// 根据角色id获取角色
	if len(r.RoleIds) == 0 {
		r.RoleIds = []uint{2} // 默认添加为普通用户角色
	}

	roles, err := isql.Role.GetRolesByIds(r.RoleIds)
	if err != nil {
		return nil, tools.NewValidatorError(fmt.Errorf("根据角色ID获取角色信息失败"))
	}

	var reqRoleSorts []int
	for _, role := range roles {
		reqRoleSorts = append(reqRoleSorts, int(role.Sort))
	}
	// 前端传来用户角色排序最小值（最高等级角色）
	reqRoleSortMin := uint(funk.MinInt(reqRoleSorts).(int))

	// 如果登录用户的角色ID为1，亦即为管理员，则直接放行，保障管理员拥有最大权限
	if currentRoleSortMin != 1 {
		// 当前用户的角色排序最小值 需要小于 前端传来的角色排序最小值（用户不能创建比自己等级高的或者相同等级的用户）
		if currentRoleSortMin >= reqRoleSortMin {
			return nil, tools.NewValidatorError(fmt.Errorf("用户不能创建比自己等级高的或者相同等级的用户"))
		}
	}
	// 获取用户将要添加的分组
	groups, err := isql.Group.GetGroupByIds(r.DepartmentId)
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("%s", "根据部门ID获取部门信息失败"+err.Error()))
	}
	var primaryPosixGroup *model.Group
	for _, g := range groups {
		ensureGroupClass(g)
		if primaryPosixGroup == nil && isPosixGroup(g) {
			primaryPosixGroup = g
		}
	}

	baseUsername := r.Username
	finalUsername := baseUsername
	suffix := 1
	for isql.User.Exist(tools.H{"username": finalUsername}) {
		finalUsername = fmt.Sprintf("%s-%d", baseUsername, suffix)
		suffix++
	}

	user := model.User{
		Username:      finalUsername,
		Password:      r.Password,
		Nickname:      r.Nickname,
		GivenName:     r.GivenName,
		Mail:          r.Mail,
		JobNumber:     r.JobNumber,
		Mobile:        r.Mobile,
		Avatar:        r.Avatar,
		PostalAddress: r.PostalAddress,
		Departments:   r.Departments,
		Position:      r.Position,
		Introduction:  r.Introduction,
		Status:        r.Status,
		Creator:       ctxUser.Username,
		DepartmentId:  tools.SliceToString(r.DepartmentId, ","),
		Source:        r.Source,
		Roles:         roles,
		UserDN:        fmt.Sprintf("uid=%s,%s", finalUsername, config.Conf.Ldap.UserDN),
		LoginShell:    r.LoginShell,
	}

	if primaryPosixGroup != nil {
		if primaryPosixGroup.GidNumber == 0 {
			return nil, tools.NewOperationError(fmt.Errorf("分配gidNumber失败: 所选posixGroup缺少gid"))
		}
		user.GidNumber = primaryPosixGroup.GidNumber
		uid, genErr := nextAvailableUID()
		if genErr != nil {
			return nil, tools.NewOperationError(fmt.Errorf("生成uidNumber失败: %v", genErr))
		}
		user.UidNumber = uid
		user.HomeDirectory = fmt.Sprintf("/home/%s/%s", primaryPosixGroup.GroupName, user.Username)
		if user.LoginShell == "" {
			user.LoginShell = getDefaultLoginShell()
		}
	}

	if user.Source == "" {
		user.Source = "platform"
	}

	err = CommonAddUser(&user, groups)
	if err != nil {
		return nil, tools.NewOperationError(fmt.Errorf("%s", "添加用户失败"+err.Error()))
	}
	return nil, nil
}

// List 数据列表
func (l UserLogic) List(c *gin.Context, req any) (data any, rspError any) {
	r, ok := req.(*request.UserListReq)
	if !ok {
		return nil, ReqAssertErr
	}
	_ = c

	users, err := isql.User.List(r)
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("%s", "获取用户列表失败："+err.Error()))
	}

	rets := make([]model.User, 0)
	for _, user := range users {
		rets = append(rets, *user)
	}
	count, err := isql.User.ListCount(r)
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("%s", "获取用户总数失败："+err.Error()))
	}

	return response.UserListRsp{
		Total: int(count),
		Users: rets,
	}, nil
}

// Update 更新数据
func (l UserLogic) Update(c *gin.Context, req any) (data any, rspError any) {
	r, ok := req.(*request.UserUpdateReq)
	if !ok {
		return nil, ReqAssertErr
	}
	_ = c

	if !isql.User.Exist(tools.H{"id": r.ID}) {
		return nil, tools.NewMySqlError(fmt.Errorf("该记录不存在"))
	}

	// 获取当前登陆用户
	ctxUser, err := isql.User.GetCurrentLoginUser(c)
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("获取当前登陆用户失败"))
	}

	// 拉取当前被操作用户的完整信息（含角色），用于字段比对
	oldFull, err := isql.User.GetUserById(uint(r.ID))
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("获取用户信息失败"))
	}
	oldRoleIds := extractRoleIDs(oldFull.Roles)
	oldDeptIds := tools.StringToSlice(oldFull.DepartmentId, ",")
	selfUpdate := int(r.ID) == int(ctxUser.ID)
	if selfUpdate {
		// 锁定所有不可变字段，防止前端误提交导致自改受限
		r.Username = oldFull.Username
		r.Nickname = oldFull.Nickname
		r.GivenName = oldFull.GivenName
		r.Mail = oldFull.Mail
		r.JobNumber = oldFull.JobNumber
		r.Mobile = oldFull.Mobile
		r.Avatar = oldFull.Avatar
		r.PostalAddress = oldFull.PostalAddress
		r.Departments = oldFull.Departments
		r.Position = oldFull.Position
		r.Introduction = oldFull.Introduction
		r.DepartmentId = oldDeptIds
		r.RoleIds = oldRoleIds
	}

	// 获取当前登陆用户角色ID集合
	var currentRoleSorts []int
	for _, role := range ctxUser.Roles {
		currentRoleSorts = append(currentRoleSorts, int(role.Sort))
	}

	// 获取将要操作的用户角色ID集合
	var reqRoleSorts []int
	roles, _ := isql.Role.GetRolesByIds(r.RoleIds)
	if len(roles) == 0 {
		return nil, tools.NewValidatorError(fmt.Errorf("根据角色ID获取角色信息失败"))
	}
	for _, role := range roles {
		reqRoleSorts = append(reqRoleSorts, int(role.Sort))
	}

	// 当前登陆用户角色排序最小值（最高等级角色）
	currentRoleSortMin := funk.MinInt(currentRoleSorts).(int)
	// 前端传来用户角色排序最小值（最高等级角色）
	reqRoleSortMin := funk.MinInt(reqRoleSorts).(int)

	// 如果登录用户的角色ID为1，亦即为管理员，则直接放行，保障管理员拥有最大权限
	if currentRoleSortMin != 1 {
		// 判断是更新自己还是更新别人,如果操作的ID与登陆用户的ID一致，则说明操作的是自己
		if int(r.ID) == int(ctxUser.ID) {
			// 不能更改自己的角色
			reqDiff, currentDiff := funk.Difference(reqRoleSorts, currentRoleSorts)
			if len(reqDiff.([]int)) > 0 || len(currentDiff.([]int)) > 0 {
				return nil, tools.NewValidatorError(fmt.Errorf("不能更改自己的角色"))
			}
		} else {
			// 如果是更新别人，操作者不能更新比自己角色等级高的或者相同等级的用户
			minRoleSorts, err := isql.User.GetUserMinRoleSortsByIds([]uint{uint(r.ID)}) // 根据userIdID获取用户角色排序最小值
			if err != nil || len(minRoleSorts) == 0 {
				return nil, tools.NewValidatorError(fmt.Errorf("根据用户ID获取用户角色排序最小值失败"))
			}
			if currentRoleSortMin >= minRoleSorts[0] || currentRoleSortMin >= reqRoleSortMin {
				return nil, tools.NewValidatorError(fmt.Errorf("用户不能更新比自己角色等级高的或者相同等级的用户"))
			}
		}
	}

	// 先获取用户信息
	oldData := new(model.User)
	err = isql.User.Find(tools.H{"id": r.ID}, oldData)
	if err != nil {
		return nil, tools.NewMySqlError(err)
	}

	// 过滤掉前端会选择到的 请选择部门信息 这个选项
	var (
		depts   string
		deptids []uint
	)
	for _, v := range strings.Split(r.Departments, ",") {
		if v != "请选择部门信息" {
			depts += v + ","
		}
	}
	for _, j := range r.DepartmentId {
		if j != 0 {
			deptids = append(deptids, j)
		}
	}

	// 拼装新的用户信息
	user := model.User{
		Model:         oldData.Model,
		Username:      r.Username,
		Nickname:      r.Nickname,
		GivenName:     r.GivenName,
		Mail:          r.Mail,
		JobNumber:     r.JobNumber,
		Mobile:        r.Mobile,
		Avatar:        r.Avatar,
		PostalAddress: r.PostalAddress,
		Departments:   depts,
		Position:      r.Position,
		Introduction:  r.Introduction,
		Creator:       ctxUser.Username,
		DepartmentId:  tools.SliceToString(deptids, ","),
		Source:        oldData.Source,
		Roles:         roles,
		UserDN:        oldData.UserDN,
		UidNumber:     oldData.UidNumber,
		GidNumber:     oldData.GidNumber,
		HomeDirectory: oldData.HomeDirectory,
		LoginShell:    r.LoginShell,
	}
	if user.LoginShell == "" {
		user.LoginShell = oldData.LoginShell
	}

	if err = CommonUpdateUser(oldData, &user, r.DepartmentId); err != nil {
		return nil, tools.NewOperationError(fmt.Errorf("%s", "更新用户失败"+err.Error()))
	}

	return nil, nil
}

// Delete 删除数据
func (l UserLogic) Delete(c *gin.Context, req any) (data any, rspError any) {
	r, ok := req.(*request.UserDeleteReq)
	if !ok {
		return nil, ReqAssertErr
	}
	_ = c

	for _, id := range r.UserIds {
		filter := tools.H{"id": int(id)}
		if !isql.User.Exist(filter) {
			return nil, tools.NewMySqlError(fmt.Errorf("有用户不存在"))
		}
	}

	// 根据用户ID获取用户角色排序最小值
	roleMinSortList, err := isql.User.GetUserMinRoleSortsByIds(r.UserIds)
	if err != nil || len(roleMinSortList) == 0 {
		return nil, tools.NewValidatorError(fmt.Errorf("根据用户ID获取用户角色排序最小值失败"))
	}

	// 获取当前登陆用户角色排序最小值（最高等级角色）以及当前用户
	minSort, ctxUser, err := isql.User.GetCurrentUserMinRoleSort(c)
	if err != nil {
		return nil, tools.NewValidatorError(fmt.Errorf("获取当前登陆用户角色排序最小值失败"))
	}

	// 不能删除自己
	if funk.Contains(r.UserIds, ctxUser.ID) {
		return nil, tools.NewValidatorError(fmt.Errorf("用户不能删除自己"))
	}

	// 不能删除比自己(登陆用户)角色排序低(等级高)的用户
	for _, sort := range roleMinSortList {
		if int(minSort) > sort {
			return nil, tools.NewValidatorError(fmt.Errorf("用户不能删除比自己角色等级高的用户"))
		}
	}

	users, err := isql.User.GetUserByIds(r.UserIds)
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("%s", "获取用户信息失败: "+err.Error()))
	}

	// 先将用户从所有LDAP分组中移除，再删除用户
	for _, user := range users {
		_ = ildap.Group.RemoveUserFromAllGroups(&user)
		err := ildap.User.Delete(user.UserDN)
		if err != nil {
			return nil, tools.NewLdapError(fmt.Errorf("%s", "在LDAP删除用户失败"+err.Error()))
		}
	}

	// 再将用户从MySQL中删除
	err = isql.User.Delete(r.UserIds)
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("%s", "在MySQL删除用户失败: "+err.Error()))
	}
	// 触发用户删除事件（异步推送 webhook）
	for _, user := range users {
		if user.OAuthConnectorID > 0 {
			userCopy := user // 避免闭包引用同一变量
			webhook.Fire(webhook.EventUserDeleted, user.OAuthConnectorID, user.OAuthProvider, &userCopy)
		}
	}
	return nil, nil
}

// ChangePwd 修改密码
func (l UserLogic) ChangePwd(c *gin.Context, req any) (data any, rspError any) {
	r, ok := req.(*request.UserChangePwdReq)
	if !ok {
		return nil, ReqAssertErr
	}
	_ = c
	// 前端传来的密码是rsa加密的,先解密
	// 密码通过RSA解密
	decodeOldPassword, err := tools.RSADecrypt([]byte(r.OldPassword), config.Conf.System.RSAPrivateBytes)
	if err != nil {
		return nil, tools.NewValidatorError(fmt.Errorf("原密码解析失败"))
	}
	decodeNewPassword, err := tools.RSADecrypt([]byte(r.NewPassword), config.Conf.System.RSAPrivateBytes)
	if err != nil {
		return nil, tools.NewValidatorError(fmt.Errorf("新密码解析失败"))
	}
	r.OldPassword = string(decodeOldPassword)
	r.NewPassword = string(decodeNewPassword)
	// 获取当前用户
	user, err := isql.User.GetCurrentLoginUser(c)
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("获取当前登陆用户失败"))
	}
	validOld := tools.NewParPasswd(user.Password) == r.OldPassword
	if !validOld {
		validOld = tools.ValidateVerificationCode(tools.PurposePasswordChange, user.Mail, r.OldPassword, true)
	}
	if !validOld {
		return nil, tools.NewValidatorError(fmt.Errorf("原密码错误或验证码无效"))
	}
	// ldap更新密码时可以直接指定用户DN和新密码即可更改成功
	err = ildap.User.ChangePwd(user.UserDN, "", r.NewPassword)
	if err != nil {
		return nil, tools.NewLdapError(fmt.Errorf("%s", "在LDAP更新密码失败"+err.Error()))
	}

	// 更新密码
	err = isql.User.ChangePwd(user.Username, tools.NewGenPasswd(r.NewPassword))
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("%s", "在MySQL更新密码失败: "+err.Error()))
	}

	return nil, nil
}

// SendPasswordChangeCode 发送修改密码验证码
func (l UserLogic) SendPasswordChangeCode(c *gin.Context, req any) (data any, rspError any) {
	_, ok := req.(*request.UserSendPasswordCodeReq)
	if !ok {
		return nil, ReqAssertErr
	}
	user, err := isql.User.GetCurrentLoginUser(c)
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("获取当前登陆用户失败"))
	}
	_, expireAt, err := tools.SendPasswordChangeCode(user.Mail)
	if err != nil {
		return nil, tools.NewLdapError(fmt.Errorf("发送验证码失败:%s", err.Error()))
	}
	return gin.H{"expireAt": expireAt.Format("2006-01-02 15:04:05")}, nil
}

// ResetPassword 重置用户密码
func (l UserLogic) ResetPassword(c *gin.Context, req any) (data any, rspError any) {
	r, ok := req.(*request.UserResetPasswordReq)
	if !ok {
		return nil, ReqAssertErr
	}
	_ = c

	// 检查用户是否存在
	if !isql.User.Exist(tools.H{"username": r.Username}) {
		return nil, tools.NewValidatorError(fmt.Errorf("用户不存在"))
	}

	// 获取用户信息
	user := new(model.User)
	err := isql.User.Find(tools.H{"username": r.Username}, user)
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("获取用户信息失败: %s", err.Error()))
	}

	// 生成随机密码
	newPassword := tools.GenerateRandomPassword()

	// 在LDAP中更新密码
	err = ildap.User.ChangePwd(user.UserDN, "", newPassword)
	if err != nil {
		return nil, tools.NewLdapError(fmt.Errorf("在LDAP更新密码失败: %s", err.Error()))
	}

	// 在MySQL中更新密码
	err = isql.User.ChangePwd(user.Username, tools.NewGenPasswd(newPassword))
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("在MySQL更新密码失败: %s", err.Error()))
	}

	// 发送密码重置通知邮件
	if err := tools.SendPasswordResetNotification(user.Username, user.Nickname, user.Mail, newPassword); err != nil {
		common.Log.Warnf("发送密码重置通知邮件失败，用户: %s, 邮箱: %s, 错误: %v", user.Username, user.Mail, err)
	}

	return map[string]string{"newPassword": newPassword}, nil
}

// ChangeUserStatus 修改用户状态
func (l UserLogic) ChangeUserStatus(c *gin.Context, req any) (data any, rspError any) {
	r, ok := req.(*request.UserChangeUserStatusReq)
	if !ok {
		return nil, ReqAssertErr
	}
	_ = c
	// 校验工作
	filter := tools.H{"id": r.ID}
	if !isql.User.Exist(filter) {
		return nil, tools.NewValidatorError(fmt.Errorf("该用户不存在"))
	}
	user := new(model.User)
	err := isql.User.Find(filter, user)
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("%s", "在MySQL查询用户失败: "+err.Error()))
	}

	if r.Status == 1 && r.Status == user.Status {
		return nil, tools.NewValidatorError(fmt.Errorf("用户已经是在职状态"))
	}
	if r.Status == 2 && r.Status == user.Status {
		return nil, tools.NewValidatorError(fmt.Errorf("用户已经是离职状态"))
	}

	// 获取当前登录用户，只有管理员才能够将用户状态改变
	// 获取当前登陆用户角色排序最小值（最高等级角色）以及当前用户
	minSort, _, err := isql.User.GetCurrentUserMinRoleSort(c)
	if err != nil {
		return nil, tools.NewValidatorError(fmt.Errorf("获取当前登陆用户角色排序最小值失败"))
	}

	if int(minSort) != 1 {
		return nil, tools.NewValidatorError(fmt.Errorf("只有管理员才能更改用户状态"))
	}

	if r.Status == 2 {
		_ = ildap.Group.RemoveUserFromAllGroups(user)
		err = ildap.User.Delete(user.UserDN)
		if err != nil {
			return nil, tools.NewLdapError(fmt.Errorf("%s", "在LDAP删除用户失败"+err.Error()))
		}
	} else {
		err = ildap.User.Add(user)
		if err != nil {
			return nil, tools.NewLdapError(fmt.Errorf("%s", "在LDAP添加用户失败"+err.Error()))
		}
	}
	err = isql.User.ChangeStatus(int(r.ID), int(r.Status))
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("%s", "在MySQL更新用户状态失败: "+err.Error()))
	}
	return nil, nil
}

// GetUserInfo 获取用户信息
func (l UserLogic) GetUserInfo(c *gin.Context, req any) (data any, rspError any) {
	r, ok := req.(*request.UserGetUserInfoReq)
	if !ok {
		return nil, ReqAssertErr
	}

	_ = c
	_ = r

	user, err := isql.User.GetCurrentLoginUser(c)
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("%s", "获取当前用户信息失败: "+err.Error()))
	}
	return user, nil
}

// IssueSSHPubKey 为当前用户签发SSH证书
func (l UserLogic) IssueSSHPubKey(c *gin.Context, req any) (data any, rspError any) {
	_, ok := req.(*request.UserIssueSSHPubKeyReq)
	if !ok {
		return nil, ReqAssertErr
	}

	if !config.Conf.System.IssueSSHPubKey {
		return nil, tools.NewValidatorError(fmt.Errorf("未开启SSH证书签发功能"))
	}

	user, err := isql.User.GetCurrentLoginUser(c)
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("获取当前登陆用户失败"))
	}

	caSigner, caPub, err := loadCASigner()
	if err != nil {
		return nil, tools.NewValidatorError(fmt.Errorf("加载CA私钥失败: %w", err))
	}

	// 生成用户密钥对
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, tools.NewOperationError(fmt.Errorf("生成用户密钥失败: %w", err))
	}

	userSigner, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		return nil, tools.NewOperationError(fmt.Errorf("创建用户签名器失败: %w", err))
	}

	if user.UidNumber == 0 || user.GidNumber == 0 {
		return nil, tools.NewRspError(http.StatusUnauthorized, fmt.Errorf("仅 posix 用户可签发 SSH 证书"))
	}
	if user.Mail == "" {
		return nil, tools.NewRspError(http.StatusUnauthorized, fmt.Errorf("用户邮箱缺失，无法签发 SSH 证书"))
	}

	serial, err := rand.Int(rand.Reader, big.NewInt(0).Lsh(big.NewInt(1), 62))
	if err != nil {
		return nil, tools.NewOperationError(fmt.Errorf("生成证书序列号失败: %w", err))
	}

	principals := []string{user.Username}
	keyID := user.Mail

	now := time.Now()
	validDays := config.Conf.CA.CertificateValidDays
	if validDays <= 0 {
		validDays = 30
	}
	validBefore := now.Add(time.Duration(validDays) * 24 * time.Hour)
	cert := &ssh.Certificate{
		Key:             userSigner.PublicKey(),
		Serial:          serial.Uint64(),
		CertType:        ssh.UserCert,
		KeyId:           keyID,
		ValidPrincipals: principals,
		ValidAfter:      uint64(now.Add(-5 * time.Minute).Unix()),
		ValidBefore:     uint64(validBefore.Unix()),
		Permissions: ssh.Permissions{Extensions: map[string]string{
			"permit-pty":              "",
			"permit-X11-forwarding":   "",
			"permit-agent-forwarding": "",
			"permit-port-forwarding":  "",
		}},
	}

	if err := cert.SignCert(rand.Reader, caSigner); err != nil {
		return nil, tools.NewOperationError(fmt.Errorf("签发证书失败: %w", err))
	}

	privPEM, err := marshalOpenSSHEd25519PrivateKey(priv, "")
	if err != nil {
		return nil, tools.NewOperationError(fmt.Errorf("序列化私钥失败: %w", err))
	}

	userPub := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(userSigner.PublicKey())))
	userCert := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(cert)))
	caPubStr := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(caPub)))

	// 生成带日期的文件名
	dateStr := now.Format("20060102")
	privFileName := fmt.Sprintf("%s-%s", user.Username, dateStr)
	certFileName := fmt.Sprintf("%s-%s-cert.pub", user.Username, dateStr)

	// 打包为 zip
	zipBuf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuf)
	addToZip := func(name, content string) error {
		fh, err := zipWriter.Create(name)
		if err != nil {
			return err
		}
		_, err = fh.Write([]byte(content))
		return err
	}

	if err := addToZip(privFileName, string(privPEM)); err != nil {
		return nil, tools.NewOperationError(fmt.Errorf("写入私钥压缩包失败: %w", err))
	}
	if err := addToZip(certFileName, userCert); err != nil {
		return nil, tools.NewOperationError(fmt.Errorf("写入证书压缩包失败: %w", err))
	}
	if err := zipWriter.Close(); err != nil {
		return nil, tools.NewOperationError(fmt.Errorf("生成压缩包失败: %w", err))
	}

	zipB64 := base64.StdEncoding.EncodeToString(zipBuf.Bytes())
	zipName := fmt.Sprintf("%s-%s.zip", user.Username, dateStr)

	return map[string]string{
		"privateKey":  string(privPEM),
		"certificate": userCert,
		"publicKey":   userPub,
		"caPublicKey": caPubStr,
		"zipBase64":   zipB64,
		"zipName":     zipName,
		"privateFile": privFileName,
		"certFile":    certFileName,
	}, nil
}

// RevokeOriginalKeypair 预留接口：用于广播/撤销旧SSH证书，目前未实现逻辑
func (l UserLogic) RevokeOriginalKeypair(c *gin.Context, req any) (data any, rspError any) {
	_, ok := req.(*request.UserRevokeOriginalKeypairReq)
	if !ok {
		return nil, ReqAssertErr
	}
	// TODO: 实现撤销旧证书的广播或Webhook逻辑
	return gin.H{"message": "revokeOriginalKeypair not implemented"}, nil
}

// marshalOpenSSHEd25519PrivateKey 序列化 ed25519 私钥为 OpenSSH 私钥格式
func marshalOpenSSHEd25519PrivateKey(priv ed25519.PrivateKey, comment string) ([]byte, error) {
	pub := priv.Public().(ed25519.PublicKey)
	pubKey, err := ssh.NewPublicKey(pub)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	writeString := func(s string) {
		_ = binary.Write(buf, binary.BigEndian, uint32(len(s)))
		buf.WriteString(s)
	}
	writeBytes := func(b []byte) {
		_ = binary.Write(buf, binary.BigEndian, uint32(len(b)))
		buf.Write(b)
	}

	// header
	buf.WriteString("openssh-key-v1\x00")
	writeString("none")                            // ciphername
	writeString("none")                            // kdfname
	writeBytes([]byte{})                           // kdfoptions
	binary.Write(buf, binary.BigEndian, uint32(1)) // number of keys
	writeBytes(pubKey.Marshal())

	// private section
	privBuf := &bytes.Buffer{}
	// random check ints
	check := make([]byte, 4)
	if _, err := rand.Read(check); err != nil {
		return nil, err
	}
	privBuf.Write(check)
	privBuf.Write(check)
	writePrivString := func(s string) {
		_ = binary.Write(privBuf, binary.BigEndian, uint32(len(s)))
		privBuf.WriteString(s)
	}
	writePrivBytes := func(b []byte) {
		_ = binary.Write(privBuf, binary.BigEndian, uint32(len(b)))
		privBuf.Write(b)
	}

	writePrivString(pubKey.Type())
	writePrivBytes(pub)
	writePrivBytes(priv)
	writePrivString(comment)

	// padding 1..n to block size 8
	padLen := 8 - (privBuf.Len() % 8)
	if padLen == 8 {
		padLen = 0
	}
	for i := 1; i <= padLen; i++ {
		privBuf.WriteByte(byte(i))
	}

	writeBytes(privBuf.Bytes())

	return pem.EncodeToMemory(&pem.Block{Type: "OPENSSH PRIVATE KEY", Bytes: buf.Bytes()}), nil
}

func loadCASigner() (ssh.Signer, ssh.PublicKey, error) {
	caPath := config.Conf.System.CAPath
	if config.Conf.CA != nil && config.Conf.CA.CAPath != "" {
		caPath = config.Conf.CA.CAPath
	}

	paths := []string{}
	if caPath != "" {
		paths = append(paths, caPath)
		base := strings.TrimSuffix(caPath, filepath.Ext(caPath))
		if base != caPath {
			paths = append(paths, base)
			paths = append(paths, base+".key", base+".pem")
		}
	}
	// 兜底尝试常见命名
	paths = append(paths, "ca", "ca.key")

	var firstErr error
	for _, p := range paths {
		if p == "" {
			continue
		}
		b, err := os.ReadFile(p)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		signer, err := ssh.ParsePrivateKey(b)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		return signer, signer.PublicKey(), nil
	}

	if firstErr == nil {
		firstErr = fmt.Errorf("未找到可用的CA私钥文件")
	}
	return nil, nil, firstErr
}

// isLoginShellOnlyChange checks whether only loginShell differs from stored values for self-update hardening.
func isLoginShellOnlyChange(r *request.UserUpdateReq, old *model.User, oldRoleIds []uint, oldDeptIds []uint) bool {
	if r == nil || old == nil {
		return false
	}
	if r.Username != old.Username ||
		r.Nickname != old.Nickname ||
		r.GivenName != old.GivenName ||
		r.Mail != old.Mail ||
		r.JobNumber != old.JobNumber ||
		r.Mobile != old.Mobile ||
		r.Avatar != old.Avatar ||
		r.PostalAddress != old.PostalAddress ||
		r.Departments != old.Departments ||
		r.Position != old.Position ||
		r.Introduction != old.Introduction {
		return false
	}
	if !equalUintSets(r.DepartmentId, oldDeptIds) {
		return false
	}
	if !equalUintSets(r.RoleIds, oldRoleIds) {
		return false
	}
	return true
}

func equalUintSets(a, b []uint) bool {
	if len(a) != len(b) {
		return false
	}
	ac := append([]uint(nil), a...)
	bc := append([]uint(nil), b...)
	sort.Slice(ac, func(i, j int) bool { return ac[i] < ac[j] })
	sort.Slice(bc, func(i, j int) bool { return bc[i] < bc[j] })
	for i := range ac {
		if ac[i] != bc[i] {
			return false
		}
	}
	return true
}

func extractRoleIDs(roles []*model.Role) []uint {
	var ids []uint
	for _, r := range roles {
		if r != nil {
			ids = append(ids, r.ID)
		}
	}
	return ids
}

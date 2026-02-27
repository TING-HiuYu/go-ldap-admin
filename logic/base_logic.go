package logic

import (
	"fmt"

	"github.com/eryajf/go-ldap-admin/config"
	"github.com/eryajf/go-ldap-admin/model"
	"github.com/eryajf/go-ldap-admin/model/request"
	"github.com/eryajf/go-ldap-admin/model/response"
	"github.com/eryajf/go-ldap-admin/public/tools"
	"github.com/eryajf/go-ldap-admin/public/version"
	"github.com/eryajf/go-ldap-admin/service/ildap"
	"github.com/eryajf/go-ldap-admin/service/isql"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BaseLogic struct{}

// SendCode 发送验证码
func (l BaseLogic) SendCode(c *gin.Context, req any) (data any, rspError any) {
	r, ok := req.(*request.BaseSendCodeReq)
	if !ok {
		return nil, ReqAssertErr
	}
	_ = c
	// 判断邮箱是否正确
	user := new(model.User)
	err := isql.User.Find(tools.H{"mail": r.Mail}, user)
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("%s", "通过邮箱查询用户失败"+err.Error()))
	}
	if user.Status != 1 || user.SyncState != 1 {
		return nil, tools.NewMySqlError(fmt.Errorf("该用户已离职或者未同步在ldap，无法重置密码，如有疑问，请联系管理员"))
	}
	err = tools.SendCode([]string{r.Mail})
	if err != nil {
		return nil, tools.NewLdapError(fmt.Errorf("%s", "邮件发送失败"+err.Error()))
	}

	return nil, nil
}

// ChangePwd 重置密码
func (l BaseLogic) ChangePwd(c *gin.Context, req any) (data any, rspError any) {
	r, ok := req.(*request.BaseChangePwdReq)
	if !ok {
		return nil, ReqAssertErr
	}
	_ = c
	// 判断邮箱是否正确
	if !isql.User.Exist(tools.H{"mail": r.Mail}) {
		return nil, tools.NewValidatorError(fmt.Errorf("邮箱不存在,请检查邮箱是否正确"))
	}
	if ok := tools.ValidateVerificationCode(tools.PurposePasswordReset, r.Mail, r.Code, true); !ok {
		return nil, tools.NewValidatorError(fmt.Errorf("验证码错误或已失效，请重新获取"))
	}

	user := new(model.User)
	err := isql.User.Find(tools.H{"mail": r.Mail}, user)
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("%s", "通过邮箱查询用户失败"+err.Error()))
	}

	newpass, err := ildap.User.NewPwd(user.Username)
	if err != nil {
		return nil, tools.NewLdapError(fmt.Errorf("%s", "LDAP生成新密码失败"+err.Error()))
	}

	err = tools.SendMail([]string{user.Mail}, newpass)
	if err != nil {
		return nil, tools.NewLdapError(fmt.Errorf("%s", "邮件发送失败"+err.Error()))
	}

	// 更新数据库密码
	err = isql.User.ChangePwd(user.Username, tools.NewGenPasswd(newpass))
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("%s", "在MySQL更新密码失败: "+err.Error()))
	}

	return nil, nil
}

// Dashboard 仪表盘
func (l BaseLogic) Dashboard(c *gin.Context, req any) (data any, rspError any) {
	_, ok := req.(*request.BaseDashboardReq)
	if !ok {
		return nil, ReqAssertErr
	}
	_ = c

	userCount, err := isql.User.Count()
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("获取用户总数失败"))
	}
	groupCount, err := isql.Group.Count()
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("获取分组总数失败"))
	}
	roleCount, err := isql.Role.Count()
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("获取角色总数失败"))
	}
	menuCount, err := isql.Menu.Count()
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("获取菜单总数失败"))
	}
	apiCount, err := isql.Api.Count()
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("获取接口总数失败"))
	}
	logCount, err := isql.OperationLog.Count()
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("获取日志总数失败"))
	}

	rst := make([]*response.DashboardList, 0)

	rst = append(rst,
		&response.DashboardList{
			DataType:  "user",
			DataName:  "用户",
			DataCount: userCount,
			Icon:      "people",
			Path:      "#/personnel/user",
		},
		&response.DashboardList{
			DataType:  "group",
			DataName:  "分组",
			DataCount: groupCount,
			Icon:      "peoples",
			Path:      "#/personnel/group",
		},
		&response.DashboardList{
			DataType:  "role",
			DataName:  "角色",
			DataCount: roleCount,
			Icon:      "eye-open",
			Path:      "#/system/role",
		},
		&response.DashboardList{
			DataType:  "menu",
			DataName:  "菜单",
			DataCount: menuCount,
			Icon:      "tree-table",
			Path:      "#/system/menu",
		},
		&response.DashboardList{
			DataType:  "api",
			DataName:  "接口",
			DataCount: apiCount,
			Icon:      "tree",
			Path:      "#/system/api",
		},
		&response.DashboardList{
			DataType:  "log",
			DataName:  "日志",
			DataCount: logCount,
			Icon:      "documentation",
			Path:      "#/log/operation-log",
		},
	)

	return rst, nil
}

// EncryptPasswd
func (l BaseLogic) EncryptPasswd(c *gin.Context, req any) (data any, rspError any) {
	r, ok := req.(*request.EncryptPasswdReq)
	if !ok {
		return nil, ReqAssertErr
	}
	_ = c

	return tools.NewGenPasswd(r.Passwd), nil
}

// DecryptPasswd
func (l BaseLogic) DecryptPasswd(c *gin.Context, req any) (data any, rspError any) {
	r, ok := req.(*request.DecryptPasswdReq)
	if !ok {
		return nil, ReqAssertErr
	}
	_ = c

	return tools.NewParPasswd(r.Passwd), nil
}

// GetConfig 获取系统配置
func (l BaseLogic) GetConfig(c *gin.Context, req any) (data any, rspError any) {
	_, ok := req.(*request.BaseConfigReq)
	if !ok {
		return nil, ReqAssertErr
	}
	_ = c

	// 安全获取配置值，防止配置段缺失导致空指针
	rsp := &response.BaseConfigRsp{}
	if config.Conf.Ldap != nil {
		rsp.LdapEnableSync = config.Conf.Ldap.EnableSync
		rsp.DefaultLoginShell = config.Conf.Ldap.DefaultLoginShell
	}
	if config.Conf.DingTalk != nil {
		rsp.DingTalkEnableSync = config.Conf.DingTalk.EnableSync
	}
	if config.Conf.FeiShu != nil {
		rsp.FeiShuEnableSync = config.Conf.FeiShu.EnableSync
	}
	if config.Conf.WeCom != nil {
		rsp.WeComEnableSync = config.Conf.WeCom.EnableSync
	}

	if rsp.DefaultLoginShell == "" {
		rsp.DefaultLoginShell = "/bin/bash"
	}

	return rsp, nil
}

// SendLoginCode 邮箱验证码登录发送验证码
func (l BaseLogic) SendLoginCode(c *gin.Context, req any) (data any, rspError any) {
	r, ok := req.(*request.BaseSendLoginCodeReq)
	if !ok {
		return nil, ReqAssertErr
	}
	user := new(model.User)
	if err := isql.User.Find(tools.H{"mail": r.Mail}, user); err != nil {
		if gorm.ErrRecordNotFound == err {
			return nil, tools.NewValidatorError(fmt.Errorf("用户不存在"))
		}
		return nil, tools.NewMySqlError(fmt.Errorf("查询用户失败: %v", err))
	}
	if user.Status != 1 {
		return nil, tools.NewValidatorError(fmt.Errorf("用户已被禁用，无法登录"))
	}
	expireAt, err := tools.SendLoginCode(user.Mail)
	if err != nil {
		return nil, tools.NewValidatorError(err)
	}
	return gin.H{"expireAt": expireAt.Format("2006-01-02 15:04:05")}, nil
}

// VerifyOtpLogin 校验邮箱验证码并返回用户
func (l BaseLogic) VerifyOtpLogin(c *gin.Context, req any) (*model.User, any) {
	r, ok := req.(*request.BaseOtpLoginReq)
	if !ok {
		return nil, ReqAssertErr
	}
	user := new(model.User)
	if err := isql.User.Find(tools.H{"mail": r.Mail}, user); err != nil {
		if gorm.ErrRecordNotFound == err {
			return nil, tools.NewValidatorError(fmt.Errorf("用户不存在"))
		}
		return nil, tools.NewMySqlError(fmt.Errorf("查询用户失败: %v", err))
	}
	if user.Status != 1 {
		return nil, tools.NewValidatorError(fmt.Errorf("用户已被禁用"))
	}
	if ok := tools.ValidateVerificationCode(tools.PurposeLogin, r.Mail, r.Code, true); !ok {
		return nil, tools.NewValidatorError(fmt.Errorf("验证码错误或已失效"))
	}
	return user, nil
}

// GetVersion 获取版本信息
func (l BaseLogic) GetVersion(c *gin.Context, req any) (data any, rspError any) {
	_, ok := req.(*request.BaseVersionReq)
	if !ok {
		return nil, ReqAssertErr
	}
	_ = c

	return version.GetVersion(), nil
}

// ResetUserPassword 为用户生成随机密码并更新到数据库和 LDAP
func (l BaseLogic) ResetUserPassword(user *model.User) (string, any) {
	newPass := tools.GenerateRandomPassword()
	// 更新 LDAP 密码
	if err := ildap.User.ChangePwd(user.UserDN, "", newPass); err != nil {
		return "", tools.NewLdapError(fmt.Errorf("重置LDAP密码失败: %v", err))
	}
	// 更新数据库密码
	if err := isql.User.ChangePwd(user.Username, tools.NewGenPasswd(newPass)); err != nil {
		return "", tools.NewMySqlError(fmt.Errorf("重置数据库密码失败: %v", err))
	}
	return newPass, nil
}

package ildap

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/eryajf/go-ldap-admin/config"
	"github.com/eryajf/go-ldap-admin/model"
	"github.com/eryajf/go-ldap-admin/public/common"
	"github.com/eryajf/go-ldap-admin/public/tools"

	ldap "github.com/go-ldap/ldap/v3"
)

type UserService struct{}

// 创建资源
func (x UserService) Add(user *model.User) error {
	add := ldap.NewAddRequest(user.UserDN, nil)
	add.Attribute("objectClass", buildUserObjectClasses(user))
	add.Attribute("cn", []string{user.Username})
	add.Attribute("sn", []string{user.Nickname})
	add.Attribute("businessCategory", []string{user.Departments})
	deptNo := normalizeNumeric(user.Position)
	add.Attribute("departmentNumber", []string{deptNo})
	add.Attribute("description", []string{user.Introduction})
	add.Attribute("displayName", []string{user.Nickname})
	add.Attribute("mail", []string{user.Mail})
	empNo := normalizeNumeric(user.JobNumber)
	add.Attribute("employeeNumber", []string{empNo})
	add.Attribute("givenName", []string{user.GivenName})
	add.Attribute("postalAddress", []string{user.PostalAddress})
	mobile := normalizeNumeric(user.Mobile)
	add.Attribute("mobile", []string{mobile})
	add.Attribute("uid", []string{user.Username})
	if user.UidNumber > 0 {
		add.Attribute("uidNumber", []string{fmt.Sprintf("%d", user.UidNumber)})
	}
	if user.GidNumber > 0 {
		add.Attribute("gidNumber", []string{fmt.Sprintf("%d", user.GidNumber)})
	}
	if user.HomeDirectory != "" {
		add.Attribute("homeDirectory", []string{user.HomeDirectory})
	}
	if user.LoginShell != "" {
		add.Attribute("loginShell", []string{user.LoginShell})
	}
	var pass string
	if config.Conf.Ldap.UserPasswordEncryptionType == "clear" {
		pass = tools.NewParPasswd(user.Password)
	} else {
		pass = tools.EncodePass([]byte(tools.NewParPasswd(user.Password)))
	}
	add.Attribute("userPassword", []string{pass})

	// 获取 LDAP 连接
	conn, err := common.GetLDAPConn()
	defer common.PutLADPConn(conn)
	if err != nil {
		return err
	}

	return conn.Add(add)
}

// Update 更新资源
func (x UserService) Update(oldusername string, user *model.User) error {
	modify := ldap.NewModifyRequest(user.UserDN, nil)
	modify.Replace("objectClass", buildUserObjectClasses(user))
	modify.Replace("cn", []string{user.Username})
	modify.Replace("sn", []string{oldusername})
	modify.Replace("businessCategory", []string{user.Departments})
	deptNo := normalizeNumeric(user.Position)
	modify.Replace("departmentNumber", []string{deptNo})
	modify.Replace("description", []string{user.Introduction})
	modify.Replace("displayName", []string{user.Nickname})
	modify.Replace("mail", []string{user.Mail})
	empNo := normalizeNumeric(user.JobNumber)
	modify.Replace("employeeNumber", []string{empNo})
	modify.Replace("givenName", []string{user.GivenName})
	modify.Replace("postalAddress", []string{user.PostalAddress})
	mobile := normalizeNumeric(user.Mobile)
	modify.Replace("mobile", []string{mobile})
	if user.UidNumber > 0 {
		modify.Replace("uidNumber", []string{fmt.Sprintf("%d", user.UidNumber)})
	}
	if user.GidNumber > 0 {
		modify.Replace("gidNumber", []string{fmt.Sprintf("%d", user.GidNumber)})
	}
	if user.HomeDirectory != "" {
		modify.Replace("homeDirectory", []string{user.HomeDirectory})
	}
	if user.LoginShell != "" {
		modify.Replace("loginShell", []string{user.LoginShell})
	}

	// 获取 LDAP 连接
	conn, err := common.GetLDAPConn()
	defer common.PutLADPConn(conn)
	if err != nil {
		return err
	}

	err = conn.Modify(modify)
	if err != nil {
		return err
	}
	if config.Conf.Ldap.UserNameModify && oldusername != user.Username {
		modifyDn := ldap.NewModifyDNRequest(fmt.Sprintf("uid=%s,%s", oldusername, config.Conf.Ldap.UserDN), fmt.Sprintf("uid=%s", user.Username), true, "")
		return conn.ModifyDN(modifyDn)
	}
	return nil
}

// normalizeNumeric ensures LDAP numeric-like attributes are valid; defaults to 0000 when empty or non-numeric.
func normalizeNumeric(val string) string {
	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return "0000"
	}
	for _, r := range trimmed {
		if r < '0' || r > '9' {
			return "0000"
		}
	}
	return trimmed
}

func (x UserService) Exist(filter map[string]any) (bool, error) {
	filter_str := ""
	for key, value := range filter {
		filter_str += fmt.Sprintf("(%s=%s)", key, value)
	}
	search_filter := fmt.Sprintf("(&(|(objectClass=inetOrgPerson)(objectClass=simpleSecurityObject))%s)", filter_str)
	// Construct query request
	searchRequest := ldap.NewSearchRequest(
		config.Conf.Ldap.BaseDN,                                     // This is basedn, we will start searching from this node.
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false, // Here several parameters are respectively scope, derefAliases, sizeLimit, timeLimit,  typesOnly
		search_filter,  // This is Filter for LDAP query
		[]string{"DN"}, // Here are the attributes returned by the query, provided as an array. If empty, all attributes are returned
		nil,
	)

	// 获取 LDAP 连接
	conn, err := common.GetLDAPConn()
	defer common.PutLADPConn(conn)
	if err != nil {
		return false, err
	}
	var sr *ldap.SearchResult
	// Search through ldap built-in search
	sr, err = conn.Search(searchRequest)
	if err != nil {
		return false, err
	}
	if len(sr.Entries) > 0 {
		return true, nil
	}
	return false, nil
}

// Delete 删除资源
func (x UserService) Delete(udn string) error {
	del := ldap.NewDelRequest(udn, nil)
	// 获取 LDAP 连接
	conn, err := common.GetLDAPConn()
	defer common.PutLADPConn(conn)
	if err != nil {
		return err
	}
	return conn.Del(del)
}

// ChangePwd 修改用户密码，此处旧密码也可以为空，ldap可以直接通过用户DN加上新密码来进行修改
func (x UserService) ChangePwd(udn, oldpasswd, newpasswd string) error {
	if config.Conf.Ldap.UserPasswordEncryptionType == "clear" {
		return updatePasswordClear(udn, newpasswd)
	}
	modifyPass := ldap.NewPasswordModifyRequest(udn, oldpasswd, newpasswd)

	// 获取 LDAP 连接
	conn, err := common.GetLDAPConn()
	defer common.PutLADPConn(conn)
	if err != nil {
		return err
	}

	_, err = conn.PasswordModify(modifyPass)
	if err != nil {
		return fmt.Errorf("password modify failed for %s, err: %v", udn, err)
	}
	return nil
}

// NewPwd 新旧密码都是空，通过管理员可以修改成功并返回新的密码
func (x UserService) NewPwd(username string) (string, error) {
	udn := fmt.Sprintf("uid=%s,%s", username, config.Conf.Ldap.UserDN)
	if username == "admin" {
		udn = config.Conf.Ldap.AdminDN
	}
	if config.Conf.Ldap.UserPasswordEncryptionType == "clear" {
		newpass := tools.GenerateRandomPassword()
		if err := updatePasswordClear(udn, newpass); err != nil {
			return "", fmt.Errorf("password modify failed for %s, err: %v", username, err)
		}
		return newpass, nil
	}
	modifyPass := ldap.NewPasswordModifyRequest(udn, "", "")

	// 获取 LDAP 连接
	conn, err := common.GetLDAPConn()
	defer common.PutLADPConn(conn)
	if err != nil {
		return "", err
	}

	newpass, err := conn.PasswordModify(modifyPass)
	if err != nil {
		return "", fmt.Errorf("password modify failed for %s, err: %v", username, err)
	}
	return newpass.GeneratedPassword, nil
}

// ListUIDNumbers 返回LDAP中已存在的uidNumber集合
func (x UserService) ListUIDNumbers() ([]uint, error) {
	searchRequest := ldap.NewSearchRequest(
		config.Conf.Ldap.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(uidNumber=*)",
		[]string{"uidNumber"},
		nil,
	)

	conn, err := common.GetLDAPConn()
	defer common.PutLADPConn(conn)
	if err != nil {
		return nil, err
	}

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	var rst []uint
	for _, entry := range sr.Entries {
		val := entry.GetAttributeValue("uidNumber")
		if val == "" {
			continue
		}
		if num, convErr := strconv.Atoi(val); convErr == nil && num > 0 {
			rst = append(rst, uint(num))
		}
	}
	return rst, nil
}

func buildUserObjectClasses(user *model.User) []string {
	classes := []string{"inetOrgPerson", "top"}
	if user.UidNumber > 0 || user.GidNumber > 0 || user.HomeDirectory != "" || user.LoginShell != "" {
		classes = append(classes, "posixAccount")
	}
	return classes
}

func updatePasswordClear(udn, newpasswd string) error {
	modify := ldap.NewModifyRequest(udn, nil)
	modify.Replace("userPassword", []string{newpasswd})

	// 获取 LDAP 连接
	conn, err := common.GetLDAPConn()
	defer common.PutLADPConn(conn)
	if err != nil {
		return err
	}

	if err := conn.Modify(modify); err != nil {
		return fmt.Errorf("password modify failed for %s, err: %v", udn, err)
	}
	return nil
}
func (x UserService) ListUserDN() (users []*model.User, err error) {
	// Construct query request
	searchRequest := ldap.NewSearchRequest(
		config.Conf.Ldap.BaseDN,                                     // This is basedn, we will start searching from this node.
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false, // Here several parameters are respectively scope, derefAliases, sizeLimit, timeLimit,  typesOnly
		"(|(objectClass=inetOrgPerson)(objectClass=simpleSecurityObject))", // This is Filter for LDAP query
		[]string{"DN"}, // Here are the attributes returned by the query, provided as an array. If empty, all attributes are returned
		nil,
	)

	// 获取 LDAP 连接
	conn, err := common.GetLDAPConn()
	defer common.PutLADPConn(conn)
	if err != nil {
		return users, err
	}
	var sr *ldap.SearchResult
	// Search through ldap built-in search
	sr, err = conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	if len(sr.Entries) > 0 {
		for _, v := range sr.Entries {
			users = append(users, &model.User{
				UserDN: v.DN,
			})
		}
	}
	return
}

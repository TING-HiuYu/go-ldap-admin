package ildap

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/eryajf/go-ldap-admin/config"
	"github.com/eryajf/go-ldap-admin/model"
	"github.com/eryajf/go-ldap-admin/public/common"

	ldap "github.com/go-ldap/ldap/v3"
)

type GroupService struct{}

// Add 添加资源
func (x GroupService) Add(g *model.Group) error { //organizationalUnit
	if g.Remark == "" {
		g.Remark = g.GroupName
	}
	add := ldap.NewAddRequest(g.GroupDN, nil)
	class := normalizeGroupClass(g.GroupClass)
	if g.GroupType == "ou" {
		add.Attribute("objectClass", []string{"organizationalUnit", "top"}) // 如果定义了 groupOfNAmes，那么必须指定member，否则报错如下：object class 'groupOfNames' requires attribute 'member'
	} else if class == "posixGroup" {
		if g.GidNumber == 0 {
			return errors.New("posixGroup requires gidNumber")
		}
		add.Attribute("objectClass", []string{"posixGroup", "top"})
		add.Attribute("gidNumber", []string{fmt.Sprintf("%d", g.GidNumber)})
	} else {
		add.Attribute("objectClass", []string{"groupOfUniqueNames", "top"})
		add.Attribute("uniqueMember", []string{config.Conf.Ldap.AdminDN}) // 默认将admin加入，避免空成员报错
	}
	add.Attribute(g.GroupType, []string{g.GroupName})
	add.Attribute("description", []string{g.Remark})

	// 获取 LDAP 连接
	conn, err := common.GetLDAPConn()
	defer common.PutLADPConn(conn)
	if err != nil {
		return err
	}

	return conn.Add(add)
}

// UpdateGroup 更新一个分组
func (x GroupService) Update(oldGroup, newGroup *model.Group) error {
	modify1 := ldap.NewModifyRequest(oldGroup.GroupDN, nil)
	modify1.Replace("description", []string{newGroup.Remark})

	// 获取 LDAP 连接
	conn, err := common.GetLDAPConn()
	defer common.PutLADPConn(conn)
	if err != nil {
		return err
	}

	err = conn.Modify(modify1)
	if err != nil {
		return err
	}
	// 如果配置文件允许修改分组名称，且分组名称发生了变化，那么执行修改分组名称
	if config.Conf.Ldap.GroupNameModify && newGroup.GroupName != oldGroup.GroupName {
		modify2 := ldap.NewModifyDNRequest(oldGroup.GroupDN, newGroup.GroupDN, true, "")
		err := conn.ModifyDN(modify2)
		if err != nil {
			return err
		}
	}
	return nil
}

// Delete 删除资源
func (x GroupService) Delete(gdn string) error {
	del := ldap.NewDelRequest(gdn, nil)

	// 获取 LDAP 连接
	conn, err := common.GetLDAPConn()
	defer common.PutLADPConn(conn)
	if err != nil {
		return err
	}

	return conn.Del(del)
}

// AddUserToGroup 添加用户到分组
func (x GroupService) AddUserToGroup(dn, udn string) error {
	return errors.New("deprecated signature: use AddUserToGroupWithMeta")
}

// DelUserFromGroup 将用户从分组删除
func (x GroupService) RemoveUserFromGroup(gdn, udn string) error {
	return errors.New("deprecated signature: use RemoveUserFromGroupWithMeta")
}

// DelUserFromGroup 将用户从分组删除
func (x GroupService) ListGroupDN() (groups []*model.Group, err error) {
	// Construct query request
	searchRequest := ldap.NewSearchRequest(
		config.Conf.Ldap.BaseDN,                                     // This is basedn, we will start searching from this node.
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false, // Here several parameters are respectively scope, derefAliases, sizeLimit, timeLimit,  typesOnly
		"(|(objectClass=organizationalUnit)(objectClass=groupOfUniqueNames)(objectClass=posixGroup))", // This is Filter for LDAP query
		[]string{"DN"}, // Here are the attributes returned by the query, provided as an array. If empty, all attributes are returned
		nil,
	)

	// 获取 LDAP 连接
	conn, err := common.GetLDAPConn()
	defer common.PutLADPConn(conn)
	if err != nil {
		return groups, err
	}
	var sr *ldap.SearchResult
	// Search through ldap built-in search
	sr, err = conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	if len(sr.Entries) > 0 {
		for _, v := range sr.Entries {
			groups = append(groups, &model.Group{
				GroupDN: v.DN,
			})
		}
	}
	return
}

// AddUserToGroupWithMeta 添加用户到分组，支持posixGroup
func (x GroupService) AddUserToGroupWithMeta(group *model.Group, user *model.User) error {
	if group == nil || user == nil {
		return errors.New("group or user is nil")
	}
	if strings.HasPrefix(group.GroupDN, "ou=") {
		return errors.New("不能添加用户到OU组织单元")
	}
	newmr := ldap.NewModifyRequest(group.GroupDN, nil)
	if normalizeGroupClass(group.GroupClass) == "posixGroup" {
		newmr.Add("memberUid", []string{user.Username})
	} else {
		newmr.Add("uniqueMember", []string{user.UserDN})
	}

	conn, err := common.GetLDAPConn()
	defer common.PutLADPConn(conn)
	if err != nil {
		return err
	}

	err = conn.Modify(newmr)
	if err != nil {
		if ldapErr, ok := err.(*ldap.Error); ok {
			if ldapErr.ResultCode == ldap.LDAPResultAttributeOrValueExists {
				return nil
			}
		}
	}
	return err
}

// RemoveUserFromGroupWithMeta 将用户从分组删除，支持posixGroup
func (x GroupService) RemoveUserFromGroupWithMeta(group *model.Group, user *model.User) error {
	if group == nil || user == nil {
		return errors.New("group or user is nil")
	}
	newmr := ldap.NewModifyRequest(group.GroupDN, nil)
	if normalizeGroupClass(group.GroupClass) == "posixGroup" {
		newmr.Delete("memberUid", []string{user.Username})
	} else {
		newmr.Delete("uniqueMember", []string{user.UserDN})
	}

	conn, err := common.GetLDAPConn()
	defer common.PutLADPConn(conn)
	if err != nil {
		return err
	}

	return conn.Modify(newmr)
}

// ListGIDNumbers 返回LDAP中已存在的gidNumber集合
func (x GroupService) ListGIDNumbers() ([]uint, error) {
	searchRequest := ldap.NewSearchRequest(
		config.Conf.Ldap.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=posixGroup)(gidNumber=*))",
		[]string{"gidNumber"},
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
		val := entry.GetAttributeValue("gidNumber")
		if val == "" {
			continue
		}
		if num, convErr := strconv.Atoi(val); convErr == nil && num > 0 {
			rst = append(rst, uint(num))
		}
	}
	return rst, nil
}

func normalizeGroupClass(cls string) string {
	if strings.EqualFold(cls, "posixGroup") {
		return "posixGroup"
	}
	return "groupOfUniqueNames"
}

// RemoveUserFromAllGroups 将用户从所有分组移除，覆盖posixGroup与groupOfUniqueNames
func (x GroupService) RemoveUserFromAllGroups(user *model.User) error {
	if user == nil {
		return errors.New("user is nil")
	}
	conn, err := common.GetLDAPConn()
	if err != nil {
		return err
	}
	defer common.PutLADPConn(conn)

	// 从posixGroup移除 memberUid
	posixFilter := fmt.Sprintf("(&(objectClass=posixGroup)(memberUid=%s))", ldap.EscapeFilter(user.Username))
	posixReq := ldap.NewSearchRequest(
		config.Conf.Ldap.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		posixFilter,
		[]string{"dn"},
		nil,
	)
	if sr, searchErr := conn.Search(posixReq); searchErr == nil {
		for _, entry := range sr.Entries {
			mr := ldap.NewModifyRequest(entry.DN, nil)
			mr.Delete("memberUid", []string{user.Username})
			_ = conn.Modify(mr)
		}
	}

	// 从groupOfUniqueNames移除 uniqueMember
	unFilter := fmt.Sprintf("(&(objectClass=groupOfUniqueNames)(uniqueMember=%s))", ldap.EscapeFilter(user.UserDN))
	unReq := ldap.NewSearchRequest(
		config.Conf.Ldap.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		unFilter,
		[]string{"dn"},
		nil,
	)
	if sr, searchErr := conn.Search(unReq); searchErr == nil {
		for _, entry := range sr.Entries {
			mr := ldap.NewModifyRequest(entry.DN, nil)
			mr.Delete("uniqueMember", []string{user.UserDN})
			_ = conn.Modify(mr)
		}
	}

	return nil
}

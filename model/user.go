package model

import (
	"fmt"
	"time"

	"github.com/alexandrevicenzi/unchained"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	UserStatusNormal          = 0
	UserStatusDisabled        = 1 // 禁用账户：禁止登录、禁止评论、禁止上传、禁止下载、禁止收藏
	UserStatusCommentLimited  = 2 // 禁止评论
	UserStatusUploadLimited   = 3 // 禁止上传
	UserStatusDownloadLimited = 4 // 禁止下载
	UserStatusFavoriteLimited = 5 // 禁止收藏
)

// 用户的公开信息字段
var UserPublicFields = []string{
	"id", "username", "signature", "status", "avatar", "realname",
	"doc_count", "follow_count", "fans_count", "favorite_count", "comment_count",
	"created_at", "updated_at", "login_at", "credit_count",
}

type User struct {
	Id            int64      `form:"id" json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id;comment:用户 id;"`
	Username      string     `form:"username" json:"username,omitempty" gorm:"column:username;type:varchar(64);size:64;index:username,unique;comment:用户名;"`
	Password      string     `form:"password" json:"password,omitempty" gorm:"column:password;type:varchar(128);size:128;comment:密码;"`
	Mobile        string     `form:"mobile" json:"mobile,omitempty" gorm:"column:mobile;type:varchar(20);size:20;index:mobile;comment:手机号;"`
	Email         string     `form:"email" json:"email,omitempty" gorm:"column:email;type:varchar(64);size:64;index:idx_email;comment:联系邮箱;"`
	Address       string     `form:"address" json:"address,omitempty" gorm:"column:address;type:varchar(255);size:255;comment:联系地址;"`
	Signature     string     `form:"signature" json:"signature,omitempty" gorm:"column:signature;type:varchar(255);size:255;comment:签名;"`
	LastLoginIp   string     `form:"last_login_ip" json:"last_login_ip,omitempty" gorm:"column:last_login_ip;type:varchar(64);size:64;comment:最后登录 ip 地址;"`
	RegisterIp    string     `form:"register_ip" json:"register_ip,omitempty" gorm:"column:register_ip;type:varchar(64);size:64;comment:注册ip;"`
	DocCount      int        `form:"doc_count" json:"doc_count,omitempty" gorm:"column:doc_count;type:int(10);default:0;comment:上传的文档数;"`
	FollowCount   int        `form:"follow_count" json:"follow_count,omitempty" gorm:"column:follow_count;type:int(10);default:0;comment:关注数;"`
	FansCount     int        `form:"fans_count" json:"fans_count,omitempty" gorm:"column:fans_count;type:int(10);default:0;comment:粉丝数;"`
	FavoriteCount int        `form:"favorite_count" json:"favorite_count,omitempty" gorm:"column:favorite_count;type:int(10);default:0;comment:收藏数;"`
	CommentCount  int        `form:"comment_count" json:"comment_count,omitempty" gorm:"column:comment_count;type:int(11);size:11;default:0;comment:评论数;"`
	CreditCount   int        `form:"credit_count" json:"credit_count,omitempty" gorm:"column:credit_count;type:int(11);size:11;default:0;comment:积分数量;"`
	ArticleCount  int        `form:"article_count" json:"article_count,omitempty" gorm:"column:article_count;type:int(11);size:11;default:0;comment:文章数量;"`
	Avatar        string     `form:"avatar" json:"avatar,omitempty" gorm:"column:avatar;type:varchar(255);size:255;comment:头像;"`
	Identity      string     `form:"identity" json:"identity,omitempty" gorm:"column:identity;type:char(18);size:18;comment:身份证号码;"`
	Realname      string     `form:"realname" json:"realname,omitempty" gorm:"column:realname;type:varchar(20);size:20;comment:身份证姓名;"`
	LoginAt       *time.Time `form:"login_at" json:"login_at,omitempty" gorm:"column:login_at;type:datetime;comment:最后登录时间;"`
	CreatedAt     *time.Time `form:"created_at" json:"created_at,omitempty" gorm:"column:created_at;type:datetime;comment:创建时间;"`
	UpdatedAt     *time.Time `form:"updated_at" json:"updated_at,omitempty" gorm:"column:updated_at;type:datetime;comment:更新时间;"`
	Remark        string     `form:"remark" json:"remark,omitempty" gorm:"column:remark;type:varchar(255);size:255;comment:备注;"`
}

// GetUserPublicFields 获取用户公开字段
func (m *DBModel) GetUserPublicFields() []string {
	return []string{"id", "username", "avatar", "signature", "doc_count", "follow_count", "fans_count",
		"favorite_count", "comment_count", "credit_count", "article_count"}
}

// CreateUser 创建User
func (m *DBModel) CreateUser(user *User, groupIds ...int64) (err error) {
	user.Password, _ = unchained.MakePassword(user.Password, unchained.GetRandomString(4), "md5")

	sess := m.db.Begin()
	defer func() {
		if err != nil {
			sess.Rollback()
		} else {
			sess.Commit()
		}
	}()

	// 1. 添加用户
	err = sess.Create(user).Error
	if err != nil {
		m.logger.Error("CreateUser", zap.Error(err))
		return
	}

	// 2. 添加用户组
	for _, groupId := range groupIds {
		group := &UserGroup{
			UserId:  user.Id,
			GroupId: groupId,
		}
		err = sess.Create(group).Error
		if err != nil {
			m.logger.Error("CreateUser", zap.Error(err))
			return
		}

		// 3. 添加用户统计
		err = sess.Model(&Group{}).Where("id = ?", groupId).Update("user_count", gorm.Expr("user_count + ?", 1)).Error
		if err != nil {
			m.logger.Error("CreateUser", zap.Error(err))
			return
		}
	}

	return
}

// UpdateUserPassword 更新User密码
func (m *DBModel) UpdateUserPassword(id interface{}, newPassword string, tx ...*gorm.DB) (err error) {
	newPassword, _ = unchained.MakePassword(newPassword, unchained.GetRandomString(4), "md5")

	user := &User{}
	sess := m.db.Model(user)
	if len(tx) > 0 {
		sess = tx[0].Model(user)
	}
	return sess.Where("id = ?", id).Update("password", newPassword).Error
}

// GetUser 根据id获取User
func (m *DBModel) GetUser(id int64, fields ...string) (user User, err error) {
	if id <= 0 {
		return user, gorm.ErrRecordNotFound
	}

	db := m.db
	fields = m.FilterValidFields(TableUser, fields...)
	if len(fields) > 0 {
		db = db.Select(fields)
	}

	err = db.Where("id = ?", id).First(&user).Error
	return
}

// GetUserByUsername(username string, fields ...string) 根据唯一索引获取User
func (m *DBModel) GetUserByUsername(username string, fields ...string) (user User, err error) {
	db := m.db

	fields = m.FilterValidFields(TableUser, fields...)
	if len(fields) > 0 {
		db = db.Select(fields)
	}

	db = db.Where("username = ?", username)

	err = db.First(&user).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.logger.Error("GetUserByUsername", zap.Error(err))
		return
	}
	return
}

func (m *DBModel) GetUserByEmail(email string, fields ...string) (user User, err error) {
	db := m.db
	fields = m.FilterValidFields(TableUser, fields...)
	if len(fields) > 0 {
		db = db.Select(fields)
	}

	db = db.Where("email = ?", email)

	err = db.First(&user).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.logger.Error("GetUserByEmail", zap.Error(err))
		return
	}
	return
}

// DeleteUser 删除数据
// TODO: 删除数据之后，存在 user_id 的关联表，需要删除对应数据，同时相关表的统计数值，也要随着减少
// TODO: 删除关联表数据，以及关联表的关联表数据，同时相关文件也一并删除掉
func (m *DBModel) DeleteUser(ids []int64) (err error) {
	sess := m.db.Begin()
	defer func() {
		if err != nil {
			sess.Rollback()
		} else {
			sess.Commit()
		}
	}()

	// id==1的用户不允许删除
	err = sess.Where("id in (?) and id != ?", ids, 1).Delete(&User{}).Error
	if err != nil {
		m.logger.Error("DeleteUser", zap.Error(err))
	}
	return
}

func (m *DBModel) initUser() (err error) {
	// 如果不存在任意用户，则初始化一个用户作为管理员
	var existUser User
	m.db.Select("id").First(&existUser)
	if existUser.Id > 0 {
		return
	}

	// 初始化一个用户
	user := &User{Username: "admin", Password: "mnt.ltd"}
	var groupId int64 = 1 // ID==1的用户组为管理员组
	return m.CreateUser(user, groupId)
}

// GetUserPermissinsByUserId 根据用户ID获取用户权限
func (m *DBModel) GetUserPermissinsByUserId(userId int64) (permissions []*Permission, err error) {
	if userId == 1 {
		// id==1的用户，拥有所有权限
		err = m.db.Find(&permissions).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			m.logger.Error("GetUserPermissinsByUserId", zap.Error(err))
		}
		return
	}

	sql := `SELECT
			p.*
		FROM 
			%s p
		LEFT JOIN 
			%s gp 
		ON
			p.id = gp.permission_id
		LEFT JOIN 
			%s ug
		ON
			ug.group_id=gp.group_id
		WHERE
			ug.user_id=?
		group by p.id
	`
	sql = fmt.Sprintf(sql, TablePermission, TableGroupPermission, TableUserGroup)
	err = m.db.Raw(sql, userId).Find(&permissions).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.logger.Error("GetUserPermissinsByUserId", zap.Error(err))
		return
	}
	err = nil
	return
}

// SetUserGroupAndPassword 设置用户组和密码
func (m *DBModel) SetUserGroupAndPassword(userId int64, groupId []int64, password ...string) (err error) {
	tx := m.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	var (
		existUsersGroups []UserGroup
		userGroups       []UserGroup
	)

	tx.Where("user_id = ?", userId).Find(&existUsersGroups)

	// 删除旧的关联用户组
	err = tx.Where("user_id = ?", userId).Delete(&UserGroup{}).Error
	if err != nil {
		m.logger.Error("SetUserGroupAndPassword", zap.Error(err))
		return
	}

	// 设置用户组统计数据
	for _, existUsersGroup := range existUsersGroups {
		err = tx.Model(&Group{}).Where("id = ?", existUsersGroup.GroupId).Update("user_count", gorm.Expr("user_count - ?", 1)).Error
		if err != nil {
			m.logger.Error("SetUserGroupAndPassword", zap.Error(err))
			return
		}
	}

	// 设置用户组统计数据
	for _, groupId := range groupId {
		userGroups = append(userGroups, UserGroup{UserId: userId, GroupId: groupId})
		err = tx.Model(&Group{}).Where("id = ?", groupId).Update("user_count", gorm.Expr("user_count + ?", 1)).Error
		if err != nil {
			m.logger.Error("SetUserGroupAndPassword", zap.Error(err))
			return
		}
	}

	if len(userGroups) > 0 {
		err = tx.Create(&userGroups).Error
		if err != nil {
			m.logger.Error("SetUserGroupAndPassword", zap.Error(err))
			return
		}
	}

	if len(password) > 0 && password[0] != "" {
		err = m.UpdateUserPassword(userId, password[0], tx)
		if err != nil {
			m.logger.Error("UpdateUserPassword", zap.Error(err))
			return
		}
	}

	return
}

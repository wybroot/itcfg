package auth

import (
	"errors"
	"fmt"

	"itcfg/server/internal/model"
	"itcfg/server/internal/repository"
)

// UserService 用户服务
type UserService struct {
	repo       *repository.UserRepo
	jwtManager *JWTManager
}

// NewUserService 创建用户服务
func NewUserService(repo *repository.UserRepo, jwtManager *JWTManager) *UserService {
	return &UserService{repo: repo, jwtManager: jwtManager}
}

// Login 用户登录
func (s *UserService) Login(username, password string) (string, *model.User, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}
	if user.Status != "active" {
		return "", nil, errors.New("用户已被禁用")
	}
	if !CheckPassword(password, user.Password) {
		return "", nil, errors.New("用户名或密码错误")
	}

	token, err := s.jwtManager.Generate(user.ID.String(), user.Username, user.Role)
	if err != nil {
		return "", nil, fmt.Errorf("生成令牌失败: %w", err)
	}
	return token, user, nil
}

// List 用户列表
func (s *UserService) List() ([]model.User, error) {
	return s.repo.List()
}

// Create 创建用户
func (s *UserService) Create(user *model.User) error {
	// 检查用户名唯一性
	existing, _ := s.repo.GetByUsername(user.Username)
	if existing != nil {
		return errors.New("用户名已存在")
	}

	// 密码哈希
	hashed, err := HashPassword(user.Password)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}
	user.Password = hashed
	return s.repo.Create(user)
}

// Update 更新用户
func (s *UserService) Update(user *model.User) error {
	existing, err := s.repo.GetByID(user.ID.String())
	if err != nil {
		return errors.New("用户不存在")
	}

	existing.Nickname = user.Nickname
	existing.Role = user.Role
	existing.Status = user.Status

	// 如果提供了新密码，更新密码
	if user.Password != "" {
		hashed, err := HashPassword(user.Password)
		if err != nil {
			return fmt.Errorf("密码加密失败: %w", err)
		}
		existing.Password = hashed
	}

	return s.repo.Update(existing)
}

// Delete 删除用户
func (s *UserService) Delete(id string) error {
	return s.repo.Delete(id)
}

// InitAdmin 初始化管理员账户（如果没有用户则创建）
func (s *UserService) InitAdmin(username, password string) error {
	existing, _ := s.repo.GetByUsername(username)
	if existing != nil {
		return nil // 已存在
	}

	hashed, err := HashPassword(password)
	if err != nil {
		return err
	}

	return s.repo.Create(&model.User{
		Username: username,
		Password: hashed,
		Nickname: "管理员",
		Role:     "admin",
		Status:   "active",
	})
}

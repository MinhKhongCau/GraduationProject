package sysadmin

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrBrokerUnavailable          = errors.New("rabbitMQ broker connection is closed")
	ErrInvalidNotificationPayload = errors.New("notification payload is invalid")
	ErrServiceNotFound            = errors.New("consultation service not found")
	ErrInvalidServiceInput        = errors.New("service name cannot be empty and price must be greater than 0")
	ErrUserNotFound               = errors.New("user account not found")
	ErrForbidden                  = errors.New("permission denied: admin role required")
)

// ─── UC-11: NOTIFICATION SERVICE ─────────────────────────────────────────────

type NotificationEvent struct {
	ID        int64
	UserID    string
	Type      string
	Title     string
	Message   string
	IsRead    bool
	CreatedAt time.Time
	ReadAt    *time.Time
}

type NotificationService struct {
	isBrokerConnected bool
	store             map[int64]*NotificationEvent
}

func NewNotificationService(brokerConnected bool) *NotificationService {
	return &NotificationService{
		isBrokerConnected: brokerConnected,
		store:             make(map[int64]*NotificationEvent),
	}
}

func (s *NotificationService) PublishAppointmentNotification(userID string, appointmentID int64) error {
	if !s.isBrokerConnected {
		return ErrBrokerUnavailable
	}
	if userID == "" || appointmentID <= 0 {
		return ErrInvalidNotificationPayload
	}
	s.store[appointmentID] = &NotificationEvent{
		ID:        appointmentID,
		UserID:    userID,
		Type:      "APPOINTMENT_CREATED",
		Title:     "Xác nhận lịch hẹn",
		Message:   "Lịch hẹn của bạn đã được xác nhận thành công.",
		IsRead:    false,
		CreatedAt: time.Now(),
	}
	return nil
}

func (s *NotificationService) PublishPaymentNotification(userID string, orderID string, amount float64) error {
	if !s.isBrokerConnected {
		return ErrBrokerUnavailable
	}
	if userID == "" || orderID == "" || amount <= 0 {
		return ErrInvalidNotificationPayload
	}
	return nil
}

func (s *NotificationService) MarkAsRead(notifID int64, userID string) error {
	n, exists := s.store[notifID]
	if !exists || n.UserID != userID {
		return errors.New("notification not found")
	}
	n.IsRead = true
	now := time.Now()
	n.ReadAt = &now
	return nil
}

// ─── UC-15: SERVICE CATALOG MANAGEMENT ────────────────────────────────────────

type ServiceItem struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Duration    int     `json:"duration"`
	Description string  `json:"description"`
	IsActive    bool    `json:"isActive"`
}

type ServiceCatalog struct {
	store map[int64]*ServiceItem
	seq   int64
}

func NewServiceCatalog() *ServiceCatalog {
	return &ServiceCatalog{
		store: make(map[int64]*ServiceItem),
		seq:   0,
	}
}

func (c *ServiceCatalog) Create(userRole, name string, price float64, duration int, desc string) (*ServiceItem, error) {
	if userRole != "ADMIN" {
		return nil, ErrForbidden
	}
	if strings.TrimSpace(name) == "" || price <= 0 {
		return nil, ErrInvalidServiceInput
	}
	c.seq++
	item := &ServiceItem{
		ID:          c.seq,
		Name:        name,
		Price:       price,
		Duration:    duration,
		Description: desc,
		IsActive:    true,
	}
	c.store[c.seq] = item
	return item, nil
}

func (c *ServiceCatalog) Update(userRole string, id int64, price float64, desc string) (*ServiceItem, error) {
	if userRole != "ADMIN" {
		return nil, ErrForbidden
	}
	item, exists := c.store[id]
	if !exists {
		return nil, ErrServiceNotFound
	}
	if price <= 0 {
		return nil, ErrInvalidServiceInput
	}
	item.Price = price
	item.Description = desc
	return item, nil
}

func (c *ServiceCatalog) Disable(userRole string, id int64) error {
	if userRole != "ADMIN" {
		return ErrForbidden
	}
	item, exists := c.store[id]
	if !exists {
		return ErrServiceNotFound
	}
	item.IsActive = false
	return nil
}

// ─── UC-16: USER ADMIN MANAGEMENT ─────────────────────────────────────────────

type UserAccount struct {
	UserID    string `json:"userId"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	IsBlocked bool   `json:"isBlocked"`
	Reason    string `json:"reason"`
}

type UserAdminService struct {
	users map[string]*UserAccount
}

func NewUserAdminService() *UserAdminService {
	return &UserAdminService{
		users: make(map[string]*UserAccount),
	}
}

func (s *UserAdminService) SeedUser(u *UserAccount) {
	s.users[u.UserID] = u
}

func (s *UserAdminService) BlockUser(adminRole, targetUserID, reason string) error {
	if adminRole != "ADMIN" {
		return ErrForbidden
	}
	u, exists := s.users[targetUserID]
	if !exists {
		return ErrUserNotFound
	}
	u.IsBlocked = true
	u.Reason = reason
	return nil
}

func (s *UserAdminService) UnblockUser(adminRole, targetUserID string) error {
	if adminRole != "ADMIN" {
		return ErrForbidden
	}
	u, exists := s.users[targetUserID]
	if !exists {
		return ErrUserNotFound
	}
	u.IsBlocked = false
	u.Reason = ""
	return nil
}

func (s *UserAdminService) UpdateRole(adminRole, targetUserID, newRole string) error {
	if adminRole != "ADMIN" {
		return ErrForbidden
	}
	u, exists := s.users[targetUserID]
	if !exists {
		return ErrUserNotFound
	}
	u.Role = newRole
	return nil
}

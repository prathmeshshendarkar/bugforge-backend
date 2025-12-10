package interfaces

type NotificationView struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	Metadata  string `json:"metadata"`
	IsRead    bool   `json:"is_read"`
	CreatedAt string `json:"created_at"`
}

type NotificationService interface {
	SendInApp(userID, title, message, metadata string) error
	SendEmail(userID, title, message string) error
	MarkAsRead(notificationID string) error
	GetUserNotifications(userID string) ([]NotificationView, error)
	MarkAllAsRead(userID string) error
}

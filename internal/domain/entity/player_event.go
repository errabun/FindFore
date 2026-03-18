package entity

type InviteStatus int32

const (
	InviteStatusPending  InviteStatus = 0
	InviteStatusAccepted InviteStatus = 1
	InviteStatusDeclined InviteStatus = 2
	InviteStatusClosed   InviteStatus = 3
)

func (s InviteStatus) String() string {
	switch s {
	case InviteStatusPending:
		return "pending"
	case InviteStatusAccepted:
		return "accepted"
	case InviteStatusDeclined:
		return "declined"
	case InviteStatusClosed:
		return "closed"
	default:
		return "unknown"
	}
}

func ParseInviteStatus(s string) (InviteStatus, bool) {
	switch s {
	case "pending":
		return InviteStatusPending, true
	case "accepted":
		return InviteStatusAccepted, true
	case "declined":
		return InviteStatusDeclined, true
	case "closed":
		return InviteStatusClosed, true
	default:
		return -1, false
	}
}

type PlayerEvent struct {
	ID           int64
	PlayerID     int64
	EventID      int64
	InviteStatus InviteStatus
}

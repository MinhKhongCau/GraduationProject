package appointment

import "time"

// Trạng thái phòng họp của lịch hẹn, tính theo giờ slot và cửa sổ vào phòng.
const (
	MeetingStatusTooEarly = "too_early" // chưa tới giờ mở phòng
	MeetingStatusOpen     = "open"      // đang trong cửa sổ vào phòng
	MeetingStatusTooLate  = "too_late"  // đã quá giờ đóng phòng
)

const (
	DefaultMeetingJoinEarly = 15 * time.Minute
	DefaultMeetingJoinLate  = 60 * time.Minute
)

// MeetingWindow: phòng họp mở từ (start - Early) tới (end + Late).
type MeetingWindow struct {
	Early time.Duration
	Late  time.Duration
}

// MeetingStatus trả too_early | open | too_late cho lịch hẹn CONFIRMED chưa kết thúc buổi tư vấn;
// rỗng khi lịch hẹn không có phòng họp (chưa thanh toán, đã hủy, đã hoàn thành).
func (w MeetingWindow) MeetingStatus(a Appointment, nowMs int64) string {
	if a.Status != AppointmentStatusConfirmed || a.SessionCompletedAt != nil || a.StartTime == 0 {
		return ""
	}
	if nowMs < a.StartTime-w.Early.Milliseconds() {
		return MeetingStatusTooEarly
	}
	end := a.EndTime
	if end < a.StartTime {
		end = a.StartTime
	}
	if nowMs > end+w.Late.Milliseconds() {
		return MeetingStatusTooLate
	}
	return MeetingStatusOpen
}

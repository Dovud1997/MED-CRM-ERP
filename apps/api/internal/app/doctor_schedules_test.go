package app

import "testing"

func TestValidScheduleDays(t *testing.T) {
	days := make([]scheduleDayRequest, 7)
	for i := range days {
		days[i] = scheduleDayRequest{Weekday: i + 1, IsWorking: i < 5, Start: "09:00", End: "18:00", BreakFrom: "13:00", BreakTo: "14:00"}
	}
	if !validScheduleDays(days) {
		t.Fatal("expected valid weekly schedule")
	}
	days[0].BreakTo = "19:00"
	if validScheduleDays(days) {
		t.Fatal("expected break outside shift to be rejected")
	}
}

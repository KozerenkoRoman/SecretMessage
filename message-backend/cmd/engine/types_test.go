package engine_test

import "time"

type mockClock struct {
	fixedTime time.Time
}

func (m *mockClock) Now() time.Time { return m.fixedTime }

type mockRNG struct{}

func (r *mockRNG) Intn(n int) int   { return 0 }
func (r *mockRNG) Float64() float64 { return 0.0 }
func (r *mockRNG) Seed(seed int64)  {}

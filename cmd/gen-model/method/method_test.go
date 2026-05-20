package method

import "testing"

func TestIsEmpty(t *testing.T) {
	var m *CommonMethod
	if !m.IsEmpty() {
		t.Error("nil receiver 应该返回 true")
	}

	m = &CommonMethod{ID: 0}
	if !m.IsEmpty() {
		t.Error("ID 为 0 应该返回 true")
	}

	m = &CommonMethod{ID: 1}
	if m.IsEmpty() {
		t.Error("ID 非 0 应该返回 false")
	}
}

package method

// method和gen不能放在同一个path下，否则会失效！！

type CommonMethod struct {
	ID int64
}

func (m *CommonMethod) IsEmpty() bool {
	if m == nil || m.ID == 0 {
		return false
	}
	return true
}

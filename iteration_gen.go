package excelizex

type Iteration interface {
	Next() bool
	Data() any
	Close() error
}

// Stream 通过调用迭代器实现
func (e *excel) Stream(i Iteration, options ...SheetOption) {
	var (
		j int
		s Sheet
	)

	for i.Next() {
		result := i.Data()

		if j == 0 {
			s = gen(result)
			for _, o := range options {
				o(&s)
			}
		}

		// todo 直接在excel中建表并流式写入

		j++
	}

	return
}

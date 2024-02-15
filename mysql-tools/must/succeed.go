package must

func Succeed(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func SucceedWithValue[T any](v T, err error) T {
	Succeed(err)
	return v
}

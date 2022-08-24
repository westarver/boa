package boa

type ValueType int

const (
	IntType ValueType = iota
	BoolType
	FloatType
	StringType
	IntSliceType
	FloatSliceType
	StringSliceType
)

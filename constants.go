package boa

<<<<<<< HEAD
type ParamType int

const (
	TypeBool ParamType = iota
	TypeString
	TypeStringSlice
	TypeInt
	TypeIntSlice
	TypeFloat
	TypeFloatSlice
	TypeTime
	TypeTimeDuration
	TypeDate
	TypeDateRange
=======
type ValueType int

const (
	IntType ValueType = iota
	BoolType
	FloatType
	StringType
	IntSliceType
	FloatSliceType
	StringSliceType
>>>>>>> ddb2b57a0cb42366fc393fe1a983ecea453ad4b8
)

package parser

type Type int

const (
	BASIC Type = iota
	POINTER
	LIST
	MAP
	ELLIPSIS
	USERDEFINED
	CHAN
	FUNC
)

type BasicType int

const (
	INT64 BasicType = iota
	BOOL
	DOUBLE
	STRING
	INTERFACE
	CONTEXT
	ERROR
)

type TypeDetail struct {
	TypeName BasicType
	UserType string
}

func (td TypeDetail) String(userdefined bool) string {
	if userdefined {
		return td.UserType
	}
	switch td.TypeName {
	case INT64:
		return "int64"
	case BOOL:
		return "bool"
	case DOUBLE:
		return "float64"
	case STRING:
		return "string"
	case INTERFACE:
		return "interface{}"
	case CONTEXT:
		return "context.Context"
	case ERROR:
		return "error"
	}

	return ""
}

type TypeInfo struct {
	BaseType Type
	ContainerType1 Type // For list element type and map key type
	Detail TypeDetail // Info about BasicType and ContainerType1
	ContainerType2 Type // map Value type
	Container2Detail TypeDetail
}

func (t TypeInfo) IsUserDefined() bool {
	return t.BaseType == USERDEFINED || (t.BaseType == LIST && t.ContainerType1 == USERDEFINED) || (t.BaseType == MAP && t.ContainerType2 == USERDEFINED)
}

func PrependPackageName(pkgName string, t TypeInfo) TypeInfo {
	t1 := t
	if t1.BaseType == USERDEFINED || (t1.BaseType == LIST && t1.ContainerType1 == USERDEFINED) {
		t1.Detail.UserType = pkgName + "." + t1.Detail.UserType
	} else if t1.BaseType == MAP && t1.ContainerType2 == USERDEFINED {
		t1.Container2Detail.UserType = pkgName + "." + t1.Container2Detail.UserType
	}
	return t1
}

func (t TypeInfo) String() string {
	switch t.BaseType {
	case BASIC:
		return t.Detail.String(false)
	case USERDEFINED:
		return t.Detail.String(true)
	case POINTER:
		userdefined := t.ContainerType1 == USERDEFINED
		return "*" + t.Detail.String(userdefined)
	case LIST:
		userdefined := t.ContainerType1 == USERDEFINED
		return "[]" + t.Detail.String(userdefined)
	case MAP:
		return "map[" + t.Detail.String(t.ContainerType1 == USERDEFINED) + "]" + t.Container2Detail.String(t.ContainerType2 == USERDEFINED)
	case ELLIPSIS:
		return "..." + t.Detail.String(t.ContainerType1 == USERDEFINED)
	}
	return ""
}

func isSameTypeDetail(d1 TypeDetail, d2 TypeDetail) bool {
	if d1.TypeName == d2.TypeName && d1.UserType == d2.UserType {
		return true
	}
	return false
}

func isSameType(t1 TypeInfo, t2 TypeInfo) bool {
	if t1.BaseType == t2.BaseType && t1.ContainerType1 == t2.ContainerType1 && t1.ContainerType2 == t2.ContainerType2 && isSameTypeDetail(t1.Detail, t2.Detail) && isSameTypeDetail(t1.Container2Detail, t2.Container2Detail) {
		return true
	}
	return false
}

func isBasic(name string) bool {
	if name == "int64" || name == "int" || name == "int32" || name == "string" || name == "float64" || name == "float32" || name == "bool" || name == "interface" || name == "context.Context" || name == "error" {
		return true
	}

	return false
}

func getTypeDetail(name string) TypeDetail {
	if name == "int64" || name == "int" || name == "int32" {
		return TypeDetail{TypeName:INT64}
	} else if name == "string" {
		return TypeDetail{TypeName: STRING}
	} else if name == "float64" || name == "float32" {
		return TypeDetail{TypeName: DOUBLE}
	} else if name == "bool" {
		return TypeDetail{TypeName: BOOL}
	} else if name == "interface" {
		return TypeDetail{TypeName: INTERFACE}
	} else if name == "context.Context" {
		return TypeDetail{TypeName: CONTEXT}
	} else if name == "error" {
		return TypeDetail{TypeName: ERROR}
	} else {
		return TypeDetail{UserType:name}
	}
}

func stringToType(name string) TypeInfo {
	if isBasic(name) {
		return TypeInfo{BaseType: BASIC, Detail: getTypeDetail(name)}
	} else {
		return TypeInfo{BaseType: USERDEFINED, Detail: getTypeDetail(name)}
	}
}

func arrayToType(name string) TypeInfo {
	tdetail := getTypeDetail(name)
	ctype := BASIC
	if !isBasic(name) {
		ctype = USERDEFINED
	}
	return TypeInfo{BaseType: LIST, ContainerType1: ctype, Detail: tdetail}
}

func ellipsisToType(name string) TypeInfo {
	tdetail := getTypeDetail(name)
	ctype := BASIC
	if !isBasic(name) {
		ctype = USERDEFINED
	}
	return TypeInfo{BaseType: ELLIPSIS, ContainerType1: ctype, Detail: tdetail}
}

func mapToType(keyName string, valName string) TypeInfo {
	kDetail := getTypeDetail(keyName)
	kType := BASIC
	if !isBasic(keyName) {
		kType = USERDEFINED
	}
	vDetail := getTypeDetail(valName)
	vType := BASIC
	if !isBasic(valName) {
		vType = USERDEFINED
	}
	return TypeInfo{BaseType: MAP, ContainerType1: kType, Detail: kDetail, ContainerType2: vType, Container2Detail: vDetail}
}

func interfaceToType() TypeInfo {
	return TypeInfo{BaseType: BASIC, Detail: TypeDetail{TypeName: INTERFACE}}
}

func pointerToType(name string) TypeInfo {
	ctype := BASIC
	if !isBasic(name) {
		ctype = USERDEFINED
	}
	tdetail := getTypeDetail(name)
	return TypeInfo{BaseType: POINTER, ContainerType1: ctype, Detail: tdetail}
}

func ctxType() TypeInfo {
	return TypeInfo{BaseType: BASIC, Detail: getTypeDetail("context.Context")}
}

func errType() TypeInfo {
	return TypeInfo{BaseType: BASIC, Detail: getTypeDetail("error")}
}

func chanToType(name string) TypeInfo {
	ctype := BASIC
	if !isBasic(name) {
		ctype = USERDEFINED
	}
	tdetail := getTypeDetail(name)
	return TypeInfo{BaseType: CHAN, ContainerType1: ctype, Detail: tdetail}
}

func funcType() TypeInfo {
	return TypeInfo{BaseType: FUNC}
}
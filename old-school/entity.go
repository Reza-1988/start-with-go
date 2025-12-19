package main

type School struct {
	Id      uint    `json:"id,omitempty"`
	Name    string  `json:"name,omitempty"`
	Classes []Class `json:"classes,omitempty"`
}

type Class struct {
	Id       uint     `json:"id,omitempty"`
	Name     string   `json:"name,omitempty"`
	SchoolId uint     `json:"school_id,omitempty"`
	Teacher  Person   `json:"teacher,omitempty"`
	Students []Person `json:"students,omitempty"`
}

type Person struct {
	Id      uint   `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Classes []uint `json:"calasses,omitempty"`
}

type Request struct {
	Method string      `json:"method,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

type Response struct {
	Status  bool        `json:"status,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type AddStudentToClassReq struct {
	StudentId uint `json:"student_id,omitempty"`
	ClassId   uint `json:"class_id,omitempty"`
}

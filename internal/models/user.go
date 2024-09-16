package models

type User struct {
	ID             string `json:"userID"`
	Password       string `json:"password"`
	ContainerID    string `json:"containerID"`
	CourseProgress int    `json:"courseProgress"`
}

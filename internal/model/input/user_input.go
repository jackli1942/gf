package input

type UserCreateInput struct {
	Username string `json:"username" v:"required|length:3,50#Username is required|Username length must be between 3 and 50"`
	Password string `json:"password" v:"required|length:6,50#Password is required|Password length must be between 6 and 50"`
	Nickname string `json:"nickname" v:"length:0,50#Nickname length cannot exceed 50"`
	Email    string `json:"email"    v:"email#Invalid email format"`
	Status   int    `json:"status"   v:"in:1,2#Invalid status value, must be 1 or 2" d:"1"`
}

type UserUpdateInput struct {
	Nickname *string `json:"nickname" v:"length:0,50#Nickname length cannot exceed 50"`
	Email    *string `json:"email"    v:"email#Invalid email format"`
	Status   *int    `json:"status"   v:"in:1,2#Invalid status value, must be 1 or 2"`
	// Password change would typically be a separate endpoint/input
}

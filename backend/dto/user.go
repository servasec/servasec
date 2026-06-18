package dto

type UpdateUserInput struct {
	Email    *string `json:"email,omitempty" binding:"omitnil,email,max=254"`
	Role     *string `json:"role,omitempty" binding:"omitnil,oneof=admin member"` // admin/member pour le PoC
	Banned   *bool   `json:"banned,omitempty"`
}

package model

type CharacterLimitError struct {
	S string
}

func (e CharacterLimitError) Error() string {
	return e.S
}

type CharacterNotFound struct {
	S string
}

func (e CharacterNotFound) Error() string {
	return e.S
}

type ParametersNotProvided struct {
	S string
}

func (e ParametersNotProvided) Error() string {
	return e.S
}

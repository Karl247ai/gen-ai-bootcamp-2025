package service

import "errors"

var (
	// Common errors
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("resource not found")

	// Word-specific errors
	ErrWordNotFound     = errors.New("word not found")
	ErrInvalidWordData  = errors.New("invalid word data")
	ErrDuplicateWord    = errors.New("duplicate word")

	// Group-specific errors
	ErrGroupNotFound     = errors.New("group not found")
	ErrInvalidGroupData  = errors.New("invalid group data")
	ErrDuplicateGroup    = errors.New("duplicate group name")

	// Word-Group relationship errors
	ErrWordAlreadyInGroup = errors.New("word already in group")
	ErrWordNotInGroup     = errors.New("word not in group")
) 
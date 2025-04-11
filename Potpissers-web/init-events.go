package main

type abstractEvent struct {
	Message string
}

var events []abstractEvent
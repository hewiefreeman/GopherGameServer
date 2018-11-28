package gopher

import (
	"errors"
	"net/http"
)

const (
	ErrorIncorrectFunction = "Incorrect function parameters or return parameters"
	ErrorServerRunning = "Cannot call when the server is running."
)

func SetStartCallback(cb interface{}) error {
	if(serverStarted){
		return errors.New(ErrorServerRunning)
	}else if callback, ok := cb.(func()); ok {
		callbacks.Start = callback
	}else{
		return errors.New(ErrorIncorrectFunction)
	}
	return nil
}

func SetPauseCallback(cb interface{}) error {
	if(serverStarted){
		return errors.New(ErrorServerRunning)
	}else if callback, ok := cb.(func()); ok {
		callbacks.Pause = callback
	}else{
		return errors.New(ErrorIncorrectFunction)
	}
	return nil
}

func SetResumeCallback(cb interface{}) error {
	if(serverStarted){
		return errors.New(ErrorServerRunning)
	}else if callback, ok := cb.(func()); ok {
		callbacks.Resume = callback
	}else{
		return errors.New(ErrorIncorrectFunction)
	}
	return nil
}

func SetStopCallback(cb interface{}) error {
	if(serverStarted){
		return errors.New(ErrorServerRunning)
	}else if callback, ok := cb.(func()); ok {
		callbacks.Stop = callback
	}else{
		return errors.New(ErrorIncorrectFunction)
	}
	return nil
}

func SetClientConnectCallback(cb interface{}) error {
	if(serverStarted){
		return errors.New(ErrorServerRunning)
	}else if callback, ok := cb.(func(*http.ResponseWriter, *http.Request)bool); ok {
		callbacks.ClientConnect = callback
	}else{
		return errors.New(ErrorIncorrectFunction)
	}
	return nil
}

func SetLoginCallback(cb interface{}) error {
	if(serverStarted){
		return errors.New(ErrorServerRunning)
	}else if callback, ok := cb.(func(string,int,map[string]interface{},map[string]interface{})bool); ok {
		callbacks.Login = callback
	}else{
		return errors.New(ErrorIncorrectFunction)
	}
	return nil
}

func SetLogoutCallback(cb interface{}) error {
	if(serverStarted){
		return errors.New(ErrorServerRunning)
	}else if callback, ok := cb.(func(string,int)); ok {
		callbacks.Logout = callback
	}else{
		return errors.New(ErrorIncorrectFunction)
	}
	return nil
}

func SetSignupCallback(cb interface{}) error {
	if(serverStarted){
		return errors.New(ErrorServerRunning)
	}else if callback, ok := cb.(func(string,int,map[string]interface{})bool); ok {
		callbacks.Signup = callback
	}else{
		return errors.New(ErrorIncorrectFunction)
	}
	return nil
}

func SetDeleteAccountCallback(cb interface{}) error {
	if(serverStarted){
		return errors.New(ErrorServerRunning)
	}else if callback, ok := cb.(func(string,int,map[string]interface{},map[string]interface{}) bool); ok {
		callbacks.DeleteAccount = callback
	}else{
		return errors.New(ErrorIncorrectFunction)
	}
	return nil
}

func SetAccountInfoChangeCallback(cb interface{}) error {
	if(serverStarted){
		return errors.New(ErrorServerRunning)
	}else if callback, ok := cb.(func(*users.User,int,map[string]interface{},map[string]interface{})bool); ok {
		callbacks.AccountInfoChange = callback
	}else{
		return errors.New(ErrorIncorrectFunction)
	}
	return nil
}

func SetPasswordChangeCallback(cb interface{}) error {
	if(serverStarted){
		return errors.New(ErrorServerRunning)
	}else if callback, ok := cb.(func(*users.User,int,map[string]interface{},map[string]interface{})bool); ok {
		callbacks.PasswordChange = callback
	}else{
		return errors.New(ErrorIncorrectFunction)
	}
	return nil
}

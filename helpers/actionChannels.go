// This package contains helpers for the inner mechanics of the Gopher Game Server.
// You can disregard this package entirely, as you will hopefully never need to use it.
package helpers

import (

)

type ActionChannel struct {
	c *chan channelAction
}

type channelAction struct {
	Action func([]interface{})[]interface{}
	Params []interface{}

	ReturnChan chan []interface{}

	Kill bool
}

func actionChannelListener(c *chan channelAction){
	for{
		value := <-*c

		//
		if(value.Kill){
			value.ReturnChan <- []interface{}{}
			close(value.ReturnChan)
			close(*c);
			break
		}

		//
		returned := value.Action(value.Params)

		value.ReturnChan <- returned

		close(value.ReturnChan)
	}
}

func NewActionChannel() *ActionChannel {
	c := make(chan channelAction);
	go actionChannelListener(&c);
	newAC := ActionChannel{c: &c};
	return &newAC;
}

func (a *ActionChannel) Execute(action func([]interface{})[]interface{}, params []interface{}) []interface{} {
	if(a.c == nil){
		return []interface{}{}
	}

	returnChan := make(chan []interface{});
	*a.c <- channelAction{	Action: action,
						Params: params,
						ReturnChan: returnChan};
	return <- returnChan;
}

func (a *ActionChannel) Kill(){
	*a.c <- channelAction{ Kill: true };
	a.c = nil;
}

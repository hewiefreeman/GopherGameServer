// This package contains helpers for the inner mechanics of the Gopher Game Server.
// You can disregard this package entirely, as you will hopefully never need to use it.
package helpers

import (
	"sync"
)

type ActionChannel struct {
	c *chan channelAction
	mux sync.Mutex
}

type channelAction struct {
	action func([]interface{})[]interface{}
	params []interface{}

	returnChan chan []interface{}

	kill bool
}

func actionChannelListener(c *chan channelAction){
	for{
		value := <-*c

		//
		if(value.kill){
			value.returnChan <- []interface{}{}
			close(value.returnChan)
			close(*c);
			break
		}

		//
		returned := value.action(value.params)

		value.returnChan <- returned

		close(value.returnChan)
	}
}

func NewActionChannel() *ActionChannel {
	c := make(chan channelAction);
	go actionChannelListener(&c);
	newAC := ActionChannel{c: &c};
	return &newAC;
}

func (a *ActionChannel) Execute(action func([]interface{})[]interface{}, params []interface{}) []interface{} {
	//
	a.mux.Lock();
	//
	if((*a).c == nil){ return []interface{}{} }
	channel := *a.c;
	//
	a.mux.Unlock();
	//
	returnChan := make(chan []interface{});
	channel <- channelAction{action: action,
						params: params,
						returnChan: returnChan};
	//
	return <- returnChan;
}

func (a *ActionChannel) Kill(){
	(*a).mux.Lock();
	//
	*a.c <- channelAction{ kill: true };
	(*a).c = nil;
	//
	(*a).mux.Unlock();
}

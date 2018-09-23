package main

//import (
//	"testing"
//)
//
//func TestTeardownThreads(t *testing.T) {
//	expectedResult := 1
//
//	volumesManagerStop := make(chan struct{})
//	volumesManagerStopped := make(chan struct{})
//
//	go func() {
//		defer close(volumesManagerStopped)
//
//		select {
//		case <-volumesManagerStop:
//			return
//		}
//	}()
//
//	result := teardownThreads(1, volumesManagerStop, volumesManagerStopped)
//
//	if result != expectedResult {
//		t.Fatalf("Expected %d, got %d", expectedResult, result)
//	}
//}

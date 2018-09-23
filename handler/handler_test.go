package handler

import ()

//func TestSetup(t *testing.T) {
//	var fakeHandler = Bivac{}
//
//	fakeHandler.Setup("noversion")
//
//	// Check Hostname
//	if fakeHandler.Hostname == "" {
//		t.Fatal("Hostname should not be nil")
//	}
//
//	// Check default Loglevel
//	if l := log.GetLevel(); l != log.InfoLevel {
//		t.Fatalf("Expected %v loglevel by default, got %v", log.InfoLevel, l)
//	}
//
//	// Check setting Loglevel
//	fakeHandler.Config.LogLevel = "debug"
//	fakeHandler.setupLoglevel()
//	if l := log.GetLevel(); l != log.DebugLevel {
//		t.Fatalf("Expected %v loglevel, got %v", log.DebugLevel, l)
//	}
//
//	// Check setting Loglevel to wrong value
//	fakeHandler.Config.LogLevel = "wrong"
//	err := fakeHandler.setupLoglevel()
//	if err == nil {
//		t.Fatal("Expected setupLoglevel to fail")
//	}
//}

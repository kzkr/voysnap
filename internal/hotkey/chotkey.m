#import <Cocoa/Cocoa.h>
#import <CoreGraphics/CoreGraphics.h>

#include "chotkey.h"

// Implemented in Go (//export voysnapHotkeyFired). Called when a clean tap of the
// target modifier key is detected.
extern void voysnapHotkeyFired(void);

static CFMachPortRef gTap = NULL;
static CFRunLoopSourceRef gSource = NULL;
static CFRunLoopRef gRunLoop = NULL;

static int64_t gKeycode = 54; // right Command
static bool gKeyDown = false; // is the target key currently held?
static bool gConsumed = false; // was it used as a modifier (another key pressed)?
static double gDownTime = 0;

// A press shorter than this with no other key counts as a "tap".
static const double kTapThreshold = 0.5;

static CGEventRef tapCallback(CGEventTapProxy proxy, CGEventType type,
                              CGEventRef event, void *refcon) {
  (void)proxy;
  (void)refcon;

  // Re-enable if macOS disables the tap (e.g. after a slow callback).
  if (type == kCGEventTapDisabledByTimeout ||
      type == kCGEventTapDisabledByUserInput) {
    if (gTap != NULL) {
      CGEventTapEnable(gTap, true);
    }
    return event;
  }

  if (type == kCGEventFlagsChanged) {
    int64_t kc = CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
    if (kc == gKeycode) {
      gKeyDown = !gKeyDown;
      if (gKeyDown) {
        gConsumed = false;
        gDownTime = CFAbsoluteTimeGetCurrent();
      } else {
        double held = CFAbsoluteTimeGetCurrent() - gDownTime;
        if (!gConsumed && held < kTapThreshold) {
          voysnapHotkeyFired();
        }
      }
    } else if (gKeyDown) {
      // Another modifier toggled while our key was held: it's a combo.
      gConsumed = true;
    }
  } else if (type == kCGEventKeyDown) {
    if (gKeyDown) {
      // Our key was held while a normal key was pressed: it's a shortcut.
      gConsumed = true;
    }
  }

  return event;
}

int voysnap_hotkey_create(int keycode) {
  gKeycode = keycode;
  CGEventMask mask =
      CGEventMaskBit(kCGEventFlagsChanged) | CGEventMaskBit(kCGEventKeyDown);
  gTap = CGEventTapCreate(kCGSessionEventTap, kCGHeadInsertEventTap,
                          kCGEventTapOptionListenOnly, mask, tapCallback, NULL);
  if (gTap == NULL) {
    return -1; // not trusted for Accessibility
  }
  gSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, gTap, 0);
  return 0;
}

void voysnap_hotkey_run(void) {
  gRunLoop = CFRunLoopGetCurrent();
  CFRunLoopAddSource(gRunLoop, gSource, kCFRunLoopCommonModes);
  CGEventTapEnable(gTap, true);
  CFRunLoopRun();
}

void voysnap_hotkey_stop(void) {
  if (gRunLoop != NULL) {
    CFRunLoopStop(gRunLoop);
  }
}

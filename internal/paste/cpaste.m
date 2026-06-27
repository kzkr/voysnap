#import <Cocoa/Cocoa.h>
#import <ApplicationServices/ApplicationServices.h>

#include "cpaste.h"

// pid of the app that was frontmost when recording started.
static pid_t gFrontmostPID = -1;

void silentrec_clipboard_set(const char *text) {
  NSPasteboard *pb = [NSPasteboard generalPasteboard];
  [pb clearContents];
  if (text != NULL) {
    NSString *s = [NSString stringWithUTF8String:text];
    [pb setString:s forType:NSPasteboardTypeString];
  }
}

void silentrec_paste(void) {
  CGEventSourceRef src =
      CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
  const CGKeyCode kVK_ANSI_V = 0x09;

  CGEventRef down = CGEventCreateKeyboardEvent(src, kVK_ANSI_V, true);
  CGEventSetFlags(down, kCGEventFlagMaskCommand);
  CGEventRef up = CGEventCreateKeyboardEvent(src, kVK_ANSI_V, false);
  CGEventSetFlags(up, kCGEventFlagMaskCommand);

  CGEventPost(kCGHIDEventTap, down);
  CGEventPost(kCGHIDEventTap, up);

  CFRelease(down);
  CFRelease(up);
  if (src != NULL) {
    CFRelease(src);
  }
}

void silentrec_remember_frontmost(void) {
  NSRunningApplication *app =
      [[NSWorkspace sharedWorkspace] frontmostApplication];
  gFrontmostPID = app ? [app processIdentifier] : -1;
}

void silentrec_restore_frontmost(void) {
  if (gFrontmostPID <= 0) {
    return;
  }
  NSRunningApplication *app =
      [NSRunningApplication runningApplicationWithProcessIdentifier:gFrontmostPID];
  [app activateWithOptions:NSApplicationActivateAllWindows];
}

bool silentrec_accessibility_trusted(bool prompt) {
  NSDictionary *opts =
      @{(id)kAXTrustedCheckOptionPrompt : @(prompt)};
  return AXIsProcessTrustedWithOptions((CFDictionaryRef)opts);
}

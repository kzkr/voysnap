#import <Cocoa/Cocoa.h>
#import <ApplicationServices/ApplicationServices.h>

#include "cpaste.h"

// State about the app that was frontmost when recording started.
static pid_t gFrontmostPID = -1;
static bool gFrontmostIsFinder = false;

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
  NSString *bid = app ? [app bundleIdentifier] : nil;
  // No frontmost app, or the Finder (which is what's "frontmost" on the desktop),
  // means there's nowhere to paste.
  gFrontmostIsFinder =
      (bid == nil) || [bid isEqualToString:@"com.apple.finder"];
}

void silentrec_restore_frontmost(void) {
  if (gFrontmostPID <= 0) {
    return;
  }
  NSRunningApplication *app =
      [NSRunningApplication runningApplicationWithProcessIdentifier:gFrontmostPID];
  [app activateWithOptions:NSApplicationActivateAllWindows];
}

// silentrec_frontmost_is_finder reports whether the app frontmost at the last
// silentrec_remember_frontmost call was the Finder/desktop (i.e. nowhere to
// paste). We use the frontmost app's bundle id rather than the Accessibility API
// because AX can't tell a pasteable app from the desktop: Electron apps (VS Code)
// expose neither a focused element nor window, while the desktop exposes both —
// so every AX signal gets these two cases backwards.
bool silentrec_frontmost_is_finder(void) { return gFrontmostIsFinder; }

bool silentrec_accessibility_trusted(bool prompt) {
  NSDictionary *opts =
      @{(id)kAXTrustedCheckOptionPrompt : @(prompt)};
  return AXIsProcessTrustedWithOptions((CFDictionaryRef)opts);
}

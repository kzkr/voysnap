#import <Cocoa/Cocoa.h>
#import <ApplicationServices/ApplicationServices.h>

#include "cpaste.h"
#include <stdlib.h>
#include <string.h>

// pid of the app that was frontmost when recording started.
static pid_t gFrontmostPID = -1;

// Whether the focused element at that moment was an editable text field.
static bool gFocusedEditable = false;

// computeFocusedEditable inspects the system-wide focused UI element and reports
// whether it accepts text input. Requires Accessibility permission; returns
// false (→ show popup) when permission is missing or nothing text-like is focused.
static bool computeFocusedEditable(void) {
  AXUIElementRef sys = AXUIElementCreateSystemWide();
  if (sys == NULL) {
    return false;
  }

  CFTypeRef focused = NULL;
  AXError err =
      AXUIElementCopyAttributeValue(sys, kAXFocusedUIElementAttribute, &focused);
  bool editable = false;

  if (err == kAXErrorSuccess && focused != NULL) {
    AXUIElementRef el = (AXUIElementRef)focused;

    CFTypeRef role = NULL;
    if (AXUIElementCopyAttributeValue(el, kAXRoleAttribute, &role) ==
            kAXErrorSuccess &&
        role != NULL) {
      if (CFGetTypeID(role) == CFStringGetTypeID()) {
        NSString *r = (NSString *)role;
        if ([r isEqualToString:(NSString *)kAXTextFieldRole] ||
            [r isEqualToString:(NSString *)kAXTextAreaRole] ||
            [r isEqualToString:(NSString *)kAXComboBoxRole]) {
          editable = true;
        }
      }
      CFRelease(role);
    }

    // Fallback: an element whose value can be set is effectively a text input.
    if (!editable) {
      Boolean settable = false;
      if (AXUIElementIsAttributeSettable(el, kAXValueAttribute, &settable) ==
              kAXErrorSuccess &&
          settable) {
        editable = true;
      }
    }

    CFRelease(focused);
  }

  CFRelease(sys);
  return editable;
}

char *silentrec_clipboard_get(void) {
  NSPasteboard *pb = [NSPasteboard generalPasteboard];
  NSString *s = [pb stringForType:NSPasteboardTypeString];
  if (s == nil) {
    return NULL;
  }
  return strdup([s UTF8String]);
}

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
  gFocusedEditable = computeFocusedEditable();
}

bool silentrec_focused_was_editable(void) { return gFocusedEditable; }

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

void silentrec_str_free(char *s) { free(s); }

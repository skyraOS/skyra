#import <AppKit/AppKit.h>
#import <CoreGraphics/CoreGraphics.h>
// Operator-only discovery. Never exposed to a model tool.
int main(void) { @autoreleasepool {
 NSArray *windows=CFBridgingRelease(CGWindowListCopyWindowInfo(kCGWindowListOptionOnScreenOnly|kCGWindowListExcludeDesktopElements,kCGNullWindowID));
 NSMutableArray *rows=[NSMutableArray array];
 for(NSDictionary *w in windows) {
  NSNumber *pid=w[(id)kCGWindowOwnerPID], *wid=w[(id)kCGWindowNumber];
  if([w[(id)kCGWindowLayer] intValue]!=0) continue;
  NSRunningApplication *app=[NSRunningApplication runningApplicationWithProcessIdentifier:pid.intValue];
  if(!app.bundleIdentifier) continue;
  [rows addObject:@{@"pid":pid,@"window_id":wid,@"bundle_id":app.bundleIdentifier,@"app":app.localizedName?:app.bundleIdentifier,@"title":w[(id)kCGWindowName]?:@""}];
 }
 NSData *data=[NSJSONSerialization dataWithJSONObject:rows options:NSJSONWritingPrettyPrinted|NSJSONWritingSortedKeys error:nil];
 [[NSFileHandle fileHandleWithStandardOutput] writeData:data];
} return 0; }

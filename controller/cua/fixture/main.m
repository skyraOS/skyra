#import <AppKit/AppKit.h>
@interface Canvas:NSView
@end
@implementation Canvas
-(void)drawRect:(NSRect)rect {
 [[NSColor whiteColor] setFill];NSRectFill(self.bounds);
 [@"SKYRA VISION\nPurple triangle: 42" drawInRect:NSMakeRect(30,120,450,100) withAttributes:@{NSFontAttributeName:[NSFont boldSystemFontOfSize:28],NSForegroundColorAttributeName:NSColor.blackColor}];
 NSBezierPath *p=[NSBezierPath bezierPath];[p moveToPoint:NSMakePoint(80,30)];[p lineToPoint:NSMakePoint(140,110)];[p lineToPoint:NSMakePoint(200,30)];[p closePath];[NSColor.systemPurpleColor setFill];[p fill];
}
@end
int main(void) { @autoreleasepool {
 NSApplication *app=NSApplication.sharedApplication;[app setActivationPolicy:NSApplicationActivationPolicyRegular];
 NSWindow *window=[[NSWindow alloc] initWithContentRect:NSMakeRect(100,100,520,280) styleMask:NSWindowStyleMaskTitled|NSWindowStyleMaskClosable backing:NSBackingStoreBuffered defer:NO];
 window.title=@"Skyra vision fixture";window.contentView=[[Canvas alloc] initWithFrame:window.contentView.bounds];[window makeKeyAndOrderFront:nil];[app activateIgnoringOtherApps:YES];[app run];
} return 0; }

// +build darwin

package screen

// #cgo CFLAGS: -x objective-c -l/System/Library/Frameworks
// #cgo LDFLAGS: -framework Cocoa
// #include <QuartzCore/QuartzCore.h>
// #include <Cocoa/Cocoa.h>
import "C"
import "unsafe"

type Image struct {
  Width int
  Height int
  BytesPerRow int
  Data []byte
  DataLength int
}

func MakeRect(x float64, y float64, w float64, h float64 ) C.CGRect{
  return C.CGRectMake(C.CGFloat(x), C.CGFloat(y), C.CGFloat(w), C.CGFloat(h))
}


func Capture(rect C.CGRect) Image{
  image := C.CGWindowListCreateImage(
   rect,
   C.kCGWindowListOptionOnScreenOnly,
   C.kCGNullWindowID,
   C.kCGWindowImageDefault,
  )

  var width = C.CGImageGetWidth(image)
  var height= C.CGImageGetHeight(image)
  var bitmapBytesPerRow  =  (width * 4)
  var bitmapByteCount    =  (bitmapBytesPerRow * height)
  var colorSpace = C.CGColorSpaceCreateDeviceRGB()

  context := C.CGBitmapContextCreate (nil,
              width,
              height,
              C.CGImageGetBitsPerComponent(image),
              bitmapBytesPerRow,
              colorSpace,
              C.kCGImageAlphaPremultipliedLast)
  
  drawingRect := C.CGRectMake(0.0, 0.0,C.CGFloat(width), C.CGFloat(height))

  C.CGContextDrawImage(context, drawingRect, image)
  
  var cbytes *C.char = (*C.char)(C.CGBitmapContextGetData(context))
  var bytes = C.GoBytes(unsafe.Pointer(cbytes),C.int(bitmapByteCount))

  defer C.CGImageRelease(image)
  defer C.CGContextRelease(context)
  defer C.CGColorSpaceRelease(colorSpace)

  return Image{
    Width: int(width),
    Height: int(height),
    BytesPerRow: int(bitmapBytesPerRow),
    Data: bytes,
    DataLength: len(bytes),
  }
 
}
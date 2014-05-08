// +build darwin

package screen

// #cgo CFLAGS: -x objective-c -l/System/Library/Frameworks
// #cgo LDFLAGS: -framework Cocoa
// #include <QuartzCore/QuartzCore.h>
// #include <Cocoa/Cocoa.h>
// void ** window_ids_to_int_array_bridge( CGWindowID *window_ids, int length){
//   CGWindowID *windows = malloc(length * sizeof(CGWindowID));
//   memcpy(windows, window_ids, length * sizeof(CGWindowID));
//   return (void **)(windows);
// }
import "C"
import "unsafe"
import cg "github.com/atomical/coregraphics"
// import "log"

// Window Image Types
// https://developer.apple.com/library/mac/documentation/Carbon/reference/CGWindow_Reference/Constants/Constants.html#//apple_ref/doc/constant_group/Window_Image_Types
const (
  KCGWindowImageDefault             = C.kCGWindowImageDefault
  KCGWindowImageBoundsIgnoreFraming = C.kCGWindowImageBoundsIgnoreFraming
  KCGWindowImageShouldBeOpaque      = C.kCGWindowImageShouldBeOpaque
  KCGWindowImageOnlyShadows         = C.kCGWindowImageOnlyShadows
  KCGWindowImageBestResolution      = C.kCGWindowImageBestResolution
  KCGWindowImageNominalResolution   = C.kCGWindowImageNominalResolution
)

type CGImageRef struct {
 Ref C.CGImageRef
 Data []byte
}

type Image struct {
  Width int
  Height int
  BytesPerRow int
  Data []byte
  DataLength int
}

type CaptureArea struct {
  Rect cg.Rect
  WindowIds []cg.CGWindowID
}

func MakeRect(x float64, y float64, w float64, h float64 ) C.CGRect{
  return C.CGRectMake(C.CGFloat(x), C.CGFloat(y), C.CGFloat(w), C.CGFloat(h))
}

func Capture( area CaptureArea ) Image {
  rect := MakeRect( area.Rect.X, area.Rect.Y, area.Rect.Width, area.Rect.Height )
  numberOfWindows := len(area.WindowIds)
  var image C.CGImageRef

  if numberOfWindows == 0 {

    image = C.CGWindowListCreateImage(
               rect,
               C.kCGWindowListOptionOnScreenOnly,
               C.kCGNullWindowID,
               C.kCGWindowImageBoundsIgnoreFraming,
              )

  } else {

    arrayRef := C.CFArrayCreate( 
                  C.kCFAllocatorDefault,
                  C.window_ids_to_int_array_bridge((*C.CGWindowID)(WindowIdSliceToCArray(area.WindowIds)), C.int(numberOfWindows)), //&ptr, //cb,
                  C.CFIndex(numberOfWindows), 
                  nil,
                  )

    // fmt.Println(area)
    // fmt.Println(rect)
    // fmt.Println("Array count:", C.CFArrayGetCount(arrayRef))
    // fmt.Println(C.CFArrayGetValueAtIndex(arrayRef, 0))
 
    image = C.CGWindowListCreateImageFromArray(
              C.CGRectNull, //rect,
              arrayRef,
              C.kCGWindowImageBoundsIgnoreFraming,
            )
  }


  var width = C.CGImageGetWidth(image)
  var height= C.CGImageGetHeight(image)
  var bitmapBytesPerRow  =  (width * 4)
  var bitmapByteCount    =  (bitmapBytesPerRow * height)
  var colorSpace = C.CGColorSpaceCreateDeviceRGB()

  context := C.CGBitmapContextCreate (nil,
              width,
              height,
              C.CGImageGetBitsPerComponent(image),
              bitmapBytesPerRow, //stride
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

func CGWindowListCreateImageFromArray( windowIds []cg.CGWindowID )  C.CGImageRef {
  numberOfWindows := len(windowIds)
  arrayRef := C.CFArrayCreate( 
                C.kCFAllocatorDefault,
                C.window_ids_to_int_array_bridge(
                  (*C.CGWindowID)(WindowIdSliceToCArray(windowIds)), 
                  C.int(numberOfWindows)),
                C.CFIndex(numberOfWindows), 
                nil)

  image := C.CGWindowListCreateImageFromArray(
            C.CGRectNull,
            arrayRef,
            C.kCGWindowImageBoundsIgnoreFraming,
          )

  return image
}




func CGImageRefToGoBytes( image C.CGImageRef )  []byte {
  data := C.CGDataProviderCopyData(C.CGImageGetDataProvider(image))
  ptr := C.CFDataGetBytePtr(data)

  defer C.CGImageRelease(image)
  defer C.CFRelease((C.CFTypeRef)(data))
  
  return C.GoBytes( unsafe.Pointer(ptr), C.int(C.CFDataGetLength(data)))
}


//   data := C.CGDataProviderCopyData(C.CGImageGetDataProvider(image))
//   ptr := C.CFDataGetBytePtr(data)
//   len := C.CFDataGetLength(data)
//   bytes := C.GoBytes(unsafe.Pointer(ptr), C.int(len) )

//   defer C.CGImageRelease(image)
//   defer C.CFRelease((C.CFTypeRef)(data))

//   return CGImageRef{ Ref: image, Data: bytes }
// }

func WindowIdSliceToCArray (byteSlice []cg.CGWindowID ) unsafe.Pointer {
       var array = unsafe.Pointer(C.calloc(C.size_t(len(byteSlice)), 1))
       var arrayptr = uintptr(array)

       for i := 0; i < len(byteSlice); i ++ {
              *(*cg.CGWindowID )(unsafe.Pointer(arrayptr)) = cg.CGWindowID(byteSlice[i])
              arrayptr ++
       }

       return array
}

import UIKit
import GLKit
import Example

class MainViewController : UIViewController, GLKViewDelegate {
    var loaded:Bool?
    
    override func viewDidLoad() {
        loaded = false
        let view = GLKView(frame: UIScreen.main.bounds)
        let context : EAGLContext? = EAGLContext(api: .openGLES2)
        view.context = context!
        view.drawableColorFormat = .RGBA8888
        view.drawableDepthFormat = .formatNone
        view.drawableStencilFormat = .format8
        view.drawableMultisample = .multisample4X
        view.delegate = self
        self.view = view
    }
    
    func glkView(_ view: GLKView, drawIn rect: CGRect) {
        if loaded == nil || !loaded! {
            loaded = true
            let scale = UIScreen.main.nativeScale
            ExampleLoadGL(Int(rect.width * scale), Int(rect.height * scale))
        }
        
        ExampleDrawFrame()
        DispatchQueue.main.async {
            view.setNeedsDisplay()
        }
    }
    
    override func touchesBegan(_ touches: Set<UITouch>, with event: UIEvent?) {
        let scale = UIScreen.main.nativeScale
        let touch = touches.first!
        let loc = touch.location(in: self.view)
        ExampleTouchEvent("down", Int(loc.x*scale), Int(loc.y*scale))
    }
    
    override func touchesMoved(_ touches: Set<UITouch>, with event: UIEvent?) {
        let scale = UIScreen.main.nativeScale
        let touch = touches.first!
        let loc = touch.location(in: self.view)
        ExampleTouchEvent("move", Int(loc.x*scale), Int(loc.y*scale))
    }
    
    override func touchesEnded(_ touches: Set<UITouch>, with event: UIEvent?) {
        let scale = UIScreen.main.nativeScale
        let touch = touches.first!
        let loc = touch.location(in: self.view)
        ExampleTouchEvent("up", Int(loc.x*scale), Int(loc.y*scale))
    }
}

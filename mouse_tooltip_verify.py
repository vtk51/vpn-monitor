import ctypes as c
from ctypes import wintypes as w
import subprocess,time,sys,json
from pathlib import Path
u=c.WinDLL('user32');u.SetProcessDPIAware()
CB=c.WINFUNCTYPE(w.BOOL,w.HWND,w.LPARAM)
u.EnumWindows.argtypes=[CB,w.LPARAM];u.GetWindowThreadProcessId.argtypes=[w.HWND,c.POINTER(w.DWORD)]
u.GetWindowTextW.argtypes=[w.HWND,w.LPWSTR,c.c_int];u.GetClassNameW.argtypes=[w.HWND,w.LPWSTR,c.c_int]
u.GetDlgItem.argtypes=[w.HWND,c.c_int];u.GetDlgItem.restype=w.HWND
u.GetWindowRect.argtypes=[w.HWND,c.POINTER(w.RECT)]
u.GetWindowLongPtrW.argtypes=[w.HWND,c.c_int];u.GetWindowLongPtrW.restype=c.c_ssize_t
u.SetForegroundWindow.argtypes=[w.HWND];u.SetCursorPos.argtypes=[c.c_int,c.c_int]
u.SendMessageW.argtypes=[w.HWND,w.UINT,w.WPARAM,w.LPARAM];u.SendMessageW.restype=c.c_ssize_t
u.PostMessageW.argtypes=[w.HWND,w.UINT,w.WPARAM,w.LPARAM]
def tx(h,cls=False):
 b=c.create_unicode_buffer(512)
 if cls:u.GetClassNameW(h,b,512)
 else:u.SendMessageW(h,13,512,c.addressof(b))
 return b.value
def wins(pid):
 out=[]
 @CB
 def cb(h,l):
  x=w.DWORD();u.GetWindowThreadProcessId(h,c.byref(x))
  if x.value==pid:out.append(h)
  return True
 u.EnumWindows(cb,0);return out
p=subprocess.Popen([sys.argv[1]])
try:
 time.sleep(.4);h=next(x for x in wins(p.pid) if tx(x)=='VPN монитор');u.SetForegroundWindow(h)
 pin=u.GetDlgItem(h,110);r=w.RECT();u.GetWindowRect(pin,c.byref(r));x=(r.left+r.right)//2;y=(r.top+r.bottom)//2
 u.SetCursorPos(x-60,y+60);time.sleep(.2);u.SetCursorPos(x,y);time.sleep(2)
 from PIL import ImageGrab
 rr=w.RECT();u.GetWindowRect(h,c.byref(rr))
 ImageGrab.grab(bbox=(rr.left,rr.top,rr.right,rr.bottom),all_screens=True).save(Path(sys.argv[1]).with_name('compact-tooltip-off.png'))
 tips=[tx(x) for x in wins(p.pid) if tx(x,True)=='tooltips_class32'];assert any('Поверх окон: выключено' in s for s in tips),tips
 u.mouse_event(2,0,0,0,0);u.mouse_event(4,0,0,0,0);time.sleep(.2);assert u.GetWindowLongPtrW(h,-20)&8
 u.SetCursorPos(x+1,y);time.sleep(1.2)
 tips2=[tx(x) for x in wins(p.pid) if tx(x,True)=='tooltips_class32'];assert any('Поверх окон: включено' in s for s in tips2),tips2
 u.mouse_event(2,0,0,0,0);u.mouse_event(4,0,0,0,0);time.sleep(.2);assert not u.GetWindowLongPtrW(h,-20)&8
 report={'actual_mouse_pin_style_toggles':True,'tooltip_off':tips,'tooltip_on':tips2,'passed':True}
 u.PostMessageW(h,0x10,0,0);p.wait(timeout=5);assert p.returncode==0
 Path(sys.argv[1]).with_name('mouse-tooltip-verification.json').write_text(json.dumps(report,ensure_ascii=False,indent=2),encoding='utf8');print(json.dumps(report,ensure_ascii=False))
finally:
 if p.poll() is None:p.kill();p.wait()

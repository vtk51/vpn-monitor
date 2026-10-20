"""Bounded real-window icon / Apply / pause smoke test (Windows + Pillow)."""
import ctypes as c
from ctypes import wintypes as w
import subprocess,time,sys,json,struct,hashlib
from pathlib import Path
from PIL import ImageGrab
u=c.WinDLL('user32',use_last_error=True);u.SetProcessDPIAware()
g=c.WinDLL('gdi32');shell=c.WinDLL('shell32')
CB=c.WINFUNCTYPE(w.BOOL,w.HWND,w.LPARAM)
u.EnumWindows.argtypes=[CB,w.LPARAM];u.EnumChildWindows.argtypes=[w.HWND,CB,w.LPARAM]
u.GetWindowThreadProcessId.argtypes=[w.HWND,c.POINTER(w.DWORD)]
u.GetClassNameW.argtypes=[w.HWND,w.LPWSTR,c.c_int]
u.GetDlgCtrlID.argtypes=[w.HWND];u.GetDlgCtrlID.restype=c.c_int
u.SendMessageW.argtypes=[w.HWND,w.UINT,w.WPARAM,w.LPARAM];u.SendMessageW.restype=c.c_ssize_t
u.PostMessageW.argtypes=[w.HWND,w.UINT,w.WPARAM,w.LPARAM]
u.GetWindowRect.argtypes=[w.HWND,c.POINTER(w.RECT)];u.SetForegroundWindow.argtypes=[w.HWND]
u.GetClassLongPtrW.argtypes=[w.HWND,c.c_int];u.GetClassLongPtrW.restype=c.c_size_t
u.LoadIconW.argtypes=[w.HINSTANCE,c.c_void_p];u.LoadIconW.restype=w.HICON
u.GetIconInfo.argtypes=[w.HICON,c.c_void_p];u.DestroyIcon.argtypes=[w.HICON]
g.GetObjectW.argtypes=[w.HGDIOBJ,c.c_int,c.c_void_p]
g.GetDIBits.argtypes=[w.HDC,w.HBITMAP,w.UINT,w.UINT,c.c_void_p,c.c_void_p,w.UINT]
g.DeleteObject.argtypes=[w.HGDIOBJ]
shell.ExtractIconExW.argtypes=[w.LPCWSTR,c.c_int,c.POINTER(w.HICON),c.POINTER(w.HICON),w.UINT]
def text(h):
 b=c.create_unicode_buffer(262144);u.SendMessageW(h,13,len(b),c.addressof(b));return b.value
def windows(pid):
 out=[]
 @CB
 def cb(h,l):
  p=w.DWORD();u.GetWindowThreadProcessId(h,c.byref(p))
  if p.value==pid:out.append(h)
  return True
 u.EnumWindows(cb,0);return out
def waitfor(fn,seconds=15):
 end=time.monotonic()+seconds
 while time.monotonic()<end:
  result=fn()
  if result:return result
  time.sleep(.1)
 raise AssertionError('Timed out')
class IconInfo(c.Structure):_fields_=[('icon',w.BOOL),('x',w.DWORD),('y',w.DWORD),('mask',w.HBITMAP),('color',w.HBITMAP)]
class Bitmap(c.Structure):_fields_=[('type',w.LONG),('width',w.LONG),('height',w.LONG),('stride',w.LONG),('planes',w.WORD),('bits',w.WORD),('ptr',c.c_void_p)]
def icon_digest(h):
 ii=IconInfo();assert u.GetIconInfo(h,c.byref(ii))
 try:
  bm=Bitmap();assert g.GetObjectW(ii.color,c.sizeof(bm),c.byref(bm))
  header=struct.pack('<IiiHHIIiiII',40,bm.width,-bm.height,1,32,0,0,0,0,0,0)+bytes(1024)
  info=c.create_string_buffer(header);pixels=c.create_string_buffer(bm.width*bm.height*4)
  g.CreateCompatibleDC.argtypes=[w.HDC];g.CreateCompatibleDC.restype=w.HDC
  g.DeleteDC.argtypes=[w.HDC]
  dc=g.CreateCompatibleDC(None)
  try:assert g.GetDIBits(dc,ii.color,0,bm.height,pixels,info,0)==bm.height
  finally:g.DeleteDC(dc)
  return {'size':[bm.width,bm.height],'sha256':hashlib.sha256(pixels.raw).hexdigest()}
 finally:g.DeleteObject(ii.color);g.DeleteObject(ii.mask)
def resources(data):
 pe=struct.unpack_from('<I',data,60)[0];assert struct.unpack_from('<H',data,pe+4)[0]==0x8664
 opt=pe+24;assert struct.unpack_from('<H',data,opt+68)[0]==2
 rva,size=struct.unpack_from('<II',data,opt+112+16)
 sections=struct.unpack_from('<H',data,pe+6)[0];st=opt+struct.unpack_from('<H',data,pe+20)[0]
 def filepos(r):
  for i in range(sections):
   vs,va,rs,ro=struct.unpack_from('<IIII',data,st+i*40+8)
   if va<=r<va+max(vs,rs):return ro+r-va
  raise AssertionError('RVA outside sections')
 root=filepos(rva);found={}
 def walk(off,path):
  named,ids=struct.unpack_from('<HH',data,root+off+12)
  for i in range(named+ids):
   name,target=struct.unpack_from('<II',data,root+off+16+i*8)
   assert not name&0x80000000
   if target&0x80000000:walk(target&0x7fffffff,path+[name])
   else:
    addr,n=struct.unpack_from('<II',data,root+target)
    found[tuple(path+[name])]=data[filepos(addr):filepos(addr)+n]
 walk(0,[])
 groups=[(k,v) for k,v in found.items() if k[0]==14];icons={k[1]:v for k,v in found.items() if k[0]==3}
 assert len(groups)==1 and len(icons)==9
 group=groups[0][1];count=struct.unpack_from('<H',group,4)[0];assert count==9
 sizes=[]
 for i in range(count):
  width,height,_,_,planes,bits,n,ident=struct.unpack_from('<BBBBHHIH',group,6+i*14)
  assert ident in icons and len(icons[ident])==n and bits==32
  sizes.append(width or 256)
 assert set([16,32,48,256])<=set(sizes)
 return {'types':[3,14],'group_id':groups[0][0][1],'frames':sizes,'icon_count':len(icons)}
exe=Path(sys.argv[1]).resolve();out=exe.parent
report={'exe':str(exe),'resources':resources(exe.read_bytes())}
p=subprocess.Popen([str(exe)],cwd=out)
try:
 h=waitfor(lambda:next((x for x in windows(p.pid) if text(x)=='VPN монитор'),None))
 small=u.SendMessageW(h,0x7f,0,0);big=u.SendMessageW(h,0x7f,1,0)
 assert small and big,'runtime WM_GETICON small/big missing'
 assert u.GetClassLongPtrW(h,-34)==small and u.GetClassLongPtrW(h,-14)==big
 default=u.LoadIconW(None,c.c_void_p(32512));default_hash=icon_digest(default)['sha256']
 a=w.HICON();b=w.HICON();extracted=shell.ExtractIconExW(str(exe),0,c.byref(a),c.byref(b),1);assert extracted>=1 and a.value and b.value
 report['shell_extracted_handles']=extracted
 try:
  report['runtime_small']=icon_digest(small);report['runtime_big']=icon_digest(big)
  assert report['runtime_big']==icon_digest(a) and report['runtime_small']==icon_digest(b),'runtime differs from shell icon'
  assert report['runtime_big']['sha256']!=default_hash,'generic default icon'
 finally:u.DestroyIcon(a);u.DestroyIcon(b)
 report['runtime_matches_shell']=True;report['class_icons_match']=True
 children={}
 @CB
 def child(ch,l):children[u.GetDlgCtrlID(ch)]=ch;return True
 u.EnumChildWindows(h,child,0)
 waitfor(lambda:text(children[109]))
 def view():return json.loads(text(children[109]))
 def click(i):u.SendMessageW(children[i],0xf5,0,0);time.sleep(.12)
 val=c.create_unicode_buffer('137');u.SendMessageW(children[101],12,0,c.addressof(val));click(102)
 assert '137 мс' in text(children[103]);report['apply']=True
 click(107);assert view()['Paused'];time.sleep(1)
 frozen=view();time.sleep(1);assert frozen['Nodes']==view()['Nodes'] and frozen['YouTubeOutages']==view()['YouTubeOutages']
 click(107);assert not view()['Paused'];report['pause_resume_counter']=True
 waitfor(lambda:view()['Nodes'][0]['Seen'])
 for win in windows(p.pid):
  cls=c.create_unicode_buffer(100);u.GetClassNameW(win,cls,100);assert cls.value!='ConsoleWindowClass'
 report['no_console']=True
 u.SetForegroundWindow(h);time.sleep(.4);r=w.RECT();u.GetWindowRect(h,c.byref(r))
 image=out/'dashboard-actual.png';ImageGrab.grab(bbox=(r.left,r.top,r.right,r.bottom),all_screens=True).save(image)
 report['screenshot']=str(image);report['small_handle']=small;report['big_handle']=big
 report['counter_observed']=view()['YouTubeOutages']
 u.PostMessageW(h,0x10,0,0);p.wait(timeout=5);assert p.returncode==0;report['exit_code']=p.returncode;report['passed']=True
 (out/'icon-verification.json').write_text(json.dumps(report,indent=2),encoding='utf8');print(json.dumps(report,indent=2))
finally:
 if p.poll() is None:p.kill();p.wait()

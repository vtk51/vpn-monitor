"""Real compact Win32 UI smoke. Windows + Python + Pillow."""
import ctypes as c
from ctypes import wintypes as w
import subprocess,time,sys,json,struct
from pathlib import Path
from PIL import ImageGrab,ImageChops
u=c.WinDLL('user32',use_last_error=True);u.SetProcessDPIAware()
CB=c.WINFUNCTYPE(w.BOOL,w.HWND,w.LPARAM)
u.EnumWindows.argtypes=[CB,w.LPARAM];u.EnumChildWindows.argtypes=[w.HWND,CB,w.LPARAM]
u.GetWindowThreadProcessId.argtypes=[w.HWND,c.POINTER(w.DWORD)]
u.GetClassNameW.argtypes=[w.HWND,w.LPWSTR,c.c_int]
u.GetDlgCtrlID.argtypes=[w.HWND];u.GetDlgCtrlID.restype=c.c_int
u.SendMessageW.argtypes=[w.HWND,w.UINT,w.WPARAM,w.LPARAM];u.SendMessageW.restype=c.c_ssize_t
u.PostMessageW.argtypes=[w.HWND,w.UINT,w.WPARAM,w.LPARAM]
u.IsWindowVisible.argtypes=[w.HWND];u.GetWindowRect.argtypes=[w.HWND,c.POINTER(w.RECT)]
u.SetForegroundWindow.argtypes=[w.HWND]
u.GetForegroundWindow.restype=w.HWND
u.GetWindowLongPtrW.argtypes=[w.HWND,c.c_int];u.GetWindowLongPtrW.restype=c.c_ssize_t
u.GetDpiForWindow.argtypes=[w.HWND];u.GetDpiForWindow.restype=w.UINT
u.SetFocus.argtypes=[w.HWND];u.IsIconic.argtypes=[w.HWND]
u.ShowWindow.argtypes=[w.HWND,c.c_int]
def text(h):
 b=c.create_unicode_buffer(262144);u.SendMessageW(h,13,len(b),c.addressof(b));return b.value
def settext(h,s):
 b=c.create_unicode_buffer(s);u.SendMessageW(h,12,0,c.addressof(b))
def windows(pid):
 out=[]
 @CB
 def cb(h,l):
  p=w.DWORD();u.GetWindowThreadProcessId(h,c.byref(p))
  if p.value==pid:out.append(h)
  return True
 u.EnumWindows(cb,0);return out
def waitfor(fn,seconds=30):
 end=time.monotonic()+seconds
 while time.monotonic()<end:
  result=fn()
  if result:return result
  time.sleep(.15)
 raise AssertionError('Timed out: '+str(fn))
def rect(h):
 r=w.RECT();u.GetWindowRect(h,c.byref(r));return[r.left,r.top,r.right,r.bottom]
exe=Path(sys.argv[1]).resolve();out=exe.parent
b=exe.read_bytes();pe=struct.unpack_from('<I',b,0x3c)[0];assert struct.unpack_from('<H',b,pe+4)[0]==0x8664;assert struct.unpack_from('<H',b,pe+24+68)[0]==2
p=subprocess.Popen([str(exe)]);report={'exe':str(exe),'pid':p.pid,'machine':'AMD64','subsystem':2}
try:
 h=waitfor(lambda:next((x for x in windows(p.pid) if text(x)=='VPN монитор'),None),10)
 children={}
 @CB
 def child(ch,l):children[u.GetDlgCtrlID(ch)]=ch;return True
 u.EnumChildWindows(h,child,0)
 assert all(i in children for i in [101,102,103,104,105,107,108,109,110,111,112,113,114])
 def click(i):u.SendMessageW(children[i],0xf5,0,0);time.sleep(.1)
 def view():return json.loads(text(children[109]))
 def apply(s):settext(children[101],s);click(102)
 def capture(name):
  u.SetForegroundWindow(h);time.sleep(.25);r=rect(h);ImageGrab.grab(bbox=tuple(r),all_screens=True).save(out/name);return r
 waitfor(lambda:text(children[109]));initial=view();assert initial['YouTubeOutages']==0
 report['initial_counter']=0;report['dpi']=u.GetDpiForWindow(h);scale=report['dpi']/96
 original=rect(h);assert original[2]-original[0]==round(380*scale) and original[3]-original[1]==round(353*scale),original
 assert not u.GetWindowLongPtrW(h,-20)&8;assert not u.IsWindowVisible(children[108]) and not u.IsWindowVisible(children[112])
 report['collapsed_rect']=original;capture('compact-initial.png')
 # Check real UI notification and OS style, not a diagnostic pinned boolean.
 foreground=u.GetForegroundWindow();click(110);assert u.GetWindowLongPtrW(h,-20)&8;assert rect(h)==original;assert u.GetForegroundWindow()==foreground
 assert 'включено' in text(children[110]);capture('compact-pinned.png')
 report['pinned_style']=u.GetWindowLongPtrW(h,-20);click(110);assert not u.GetWindowLongPtrW(h,-20)&8;assert rect(h)==original
 report['unpinned_style']=u.GetWindowLongPtrW(h,-20);report['pin_position_size_focus_unchanged']=True
 click(111);expanded=rect(h);assert expanded[:2]==original[:2];assert expanded[2]==original[2];assert expanded[3]-expanded[1]==round(537*scale)
 assert u.IsWindowVisible(children[108]) and u.IsWindowVisible(children[112]);report['expanded_rect']=expanded
 capture('compact-expanded-initial.png');click(111);assert rect(h)==original and not u.IsWindowVisible(children[112]);report['journal_expand_collapse']=True
 # Visible controls stay inside actual client/window and do not collide.
 rects={str(i):rect(children[i]) for i in [101,102,107,110,111,113,114]}
 for i,r in rects.items():assert r[0]>=original[0] and r[1]>=original[1] and r[2]<=original[2] and r[3]<=original[3],(i,r)
 for i,a in rects.items():
  for j,b in rects.items():
   if int(i)<int(j):assert not(a[0]<b[2] and b[0]<a[2] and a[1]<b[3] and b[1]<a[3]),(i,j)
 report['control_rects']=rects
 apply('137');assert '137 мс' in text(children[103])
 for value in ['0','60001','abc','-1','1.5']:
  apply(value);assert '1 до 60000' in text(children[103]) and 'Действующий таймаут: 137 мс' in text(children[103])
 report['invalid_rejected']=['0','60001','abc','-1','1.5']
 for value in ['1','60000']:
  apply(value);assert f'Действующий таймаут: {value} мс' in text(children[103])
 report['boundaries_accepted']=[1,60000];apply('1200')
 waitfor(lambda:text(children[104]).count('таймаут 1200 мс')==5,40);report['rows_live']=text(children[104])
 waitfor(lambda:all(n['Seen'] for n in view()['Nodes']),30);capture('compact-unpinned.png')
 assert not any('Проверка   ·' in x for x in view()['Journal']);report['default_cycle_logs_absent']=True
 apply('1');waitfor(lambda:not view()['Nodes'][3]['OK'],25);failed=view();assert failed['YouTubeOutages']>=1;report['real_outage_view']=failed
 click(107);assert view()['Paused'];time.sleep(.4);capture('compact-outage-paused.png')
 frozen=view();time.sleep(2);after=view();assert frozen['Nodes']==after['Nodes'] and frozen['Cycles']==after['Cycles'] and frozen['YouTubeOutages']==after['YouTubeOutages'];report['pause_freezes_probes_counter']=True
 click(107);assert not view()['Paused'];outage_at=failed['Nodes'][3]['At'];waitfor(lambda:view()['Nodes'][3]['At']!=outage_at,20)
 consecutive=view()
 if not consecutive['Nodes'][3]['OK']:assert consecutive['YouTubeOutages']==failed['YouTubeOutages'];report['consecutive_failure_counter_stable']=True
 else:report['consecutive_failure_counter_stable']='real probe recovered'
 apply('1200');waitfor(lambda:view()['Nodes'][3]['OK'],35);report['recovery_observed']=True
 recovered=view()['YouTubeOutages'];apply('1');waitfor(lambda:not view()['Nodes'][3]['OK'],25);assert view()['YouTubeOutages']==recovered+1;report['second_outage_count']=view()['YouTubeOutages']
 apply('1200');waitfor(lambda:view()['Nodes'][3]['OK'],35)
 click(111);click(108);assert view()['Verbose'];waitfor(lambda:any('Проверка   ·' in x for x in view()['Journal']),20);click(108);assert not view()['Verbose'];report['verbosity_works']=True
 # Pause during scroll to avoid automatic follow on new events.
 click(107);time.sleep(.4);capture('compact-expanded.png');j=children[112];before=u.SendMessageW(j,0xce,0,0);before_img=ImageGrab.grab(bbox=tuple(rect(j)),all_screens=True)
 u.SendMessageW(j,0x115,6,0);time.sleep(.2);top=u.SendMessageW(j,0xce,0,0);after_img=ImageGrab.grab(bbox=tuple(rect(j)),all_screens=True)
 assert before>top and ImageChops.difference(before_img,after_img).getbbox(),(before,top);report['journal_scroll_first_lines']=[before,top];capture('compact-journal-scrolled.png')
 click(107);click(111);capture('compact-final.png')
 report['real_chart_samples']=view()['Samples'];assert len(report['real_chart_samples'])>=2
 for win in windows(p.pid):
  cls=c.create_unicode_buffer(100);u.GetClassNameW(win,cls,100);assert cls.value!='ConsoleWindowClass'
 report['no_console']=True
 log=out/'vpn-diag.log';lines=log.read_text(encoding='utf8').splitlines();assert lines and all('ВНИМАНИЕ: Сбой YouTube' in line for line in lines);report['failure_only_disk_log']=True
 click(113);assert u.IsIconic(h);u.ShowWindow(h,9);time.sleep(.2);assert not u.IsIconic(h);report['minimize_restore']=True
 # Actual titlebar close button.
 click(114);p.wait(timeout=5);assert p.returncode==0;report['exit_code']=p.returncode;report['passed']=True
 (out/'compact-verification.json').write_text(json.dumps(report,ensure_ascii=False,indent=2),encoding='utf8');print(json.dumps(report,ensure_ascii=False,indent=2))
finally:
 if p.poll() is None:p.kill();p.wait()

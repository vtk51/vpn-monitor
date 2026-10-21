"""Bounded real Windows settings QA. Only run against an isolated EXE copy.
Creates vpn-monitor.json/log beside EXE; restores defaults and removes JSON on exit.
"""
import ctypes as c
from ctypes import wintypes as w
import json,sys,time,subprocess,struct,socket,threading
from pathlib import Path
from PIL import ImageGrab
u=c.WinDLL('user32',use_last_error=True);u.SetProcessDPIAware()
CB=c.WINFUNCTYPE(w.BOOL,w.HWND,w.LPARAM)
for n,args,ret in [('EnumWindows',[CB,w.LPARAM],w.BOOL),('EnumChildWindows',[w.HWND,CB,w.LPARAM],w.BOOL),('GetWindowThreadProcessId',[w.HWND,c.POINTER(w.DWORD)],w.DWORD),('GetDlgCtrlID',[w.HWND],c.c_int),('SendMessageW',[w.HWND,w.UINT,w.WPARAM,w.LPARAM],c.c_ssize_t),('GetWindowRect',[w.HWND,c.POINTER(w.RECT)],w.BOOL),('GetWindowLongPtrW',[w.HWND,c.c_int],c.c_ssize_t),('IsWindowEnabled',[w.HWND],w.BOOL),('IsWindowVisible',[w.HWND],w.BOOL),('SetForegroundWindow',[w.HWND],w.BOOL),('GetClassNameW',[w.HWND,w.LPWSTR,c.c_int],c.c_int),('GetDpiForWindow',[w.HWND],w.UINT)]:
 f=getattr(u,n);f.argtypes=args;f.restype=ret

def text(h):
 b=c.create_unicode_buffer(262144);u.SendMessageW(h,13,len(b),c.addressof(b));return b.value
def settext(h,s):
 b=c.create_unicode_buffer(s);u.SendMessageW(h,12,0,c.addressof(b))
def wins(pid):
 out=[]
 @CB
 def cb(h,l):
  p=w.DWORD();u.GetWindowThreadProcessId(h,c.byref(p))
  if p.value==pid:out.append(h)
  return True
 u.EnumWindows(cb,0);return out
def children(h):
 out={}
 @CB
 def cb(ch,l):out[u.GetDlgCtrlID(ch)]=ch;return True
 u.EnumChildWindows(h,cb,0);return out
def wait(fn,seconds=35):
 end=time.monotonic()+seconds
 while time.monotonic()<end:
  x=fn()
  if x:return x
  time.sleep(.1)
 raise AssertionError('Timed out '+str(fn))
def click(ch):u.SendMessageW(ch,0xf5,0,0);time.sleep(.12)
def rect(h):
 r=w.RECT();u.GetWindowRect(h,c.byref(r));return [r.left,r.top,r.right,r.bottom]
exe=Path(sys.argv[1]).resolve();out=exe.parent/'qa';out.mkdir(exist_ok=True);cfg=exe.parent/'vpn-monitor.json'
assert not cfg.exists(),'Refusing personal config overwrite'
report={'exe':str(exe),'checks':{}};p=None
sock=socket.socket(socket.AF_INET,socket.SOCK_DGRAM);sock.bind(('127.0.0.1',0));sock.settimeout(.2);queries=[];stop=threading.Event()
def dns():
 while not stop.is_set():
  try:
   data,addr=sock.recvfrom(2048);i=12;labels=[]
   while data[i]:n=data[i];labels.append(data[i+1:i+1+n].decode());i+=n+1
   qend=i+5;queries.append('.'.join(labels));reply=data[:2]+b'\x81\x80\x00\x01\x00\x01\x00\x00\x00\x00'+data[12:qend]+b'\xc0\x0c\x00\x01\x00\x01\x00\x00\x00\x3c\x00\x04\x7f\x00\x00\x01';sock.sendto(reply,addr)
  except socket.timeout:pass
  except OSError:break
threading.Thread(target=dns,daemon=True).start()
def launch():
 global p,h,ch
 p=subprocess.Popen([str(exe)]);h=wait(lambda:next((x for x in wins(p.pid) if text(x)=='VPN монитор'),None),10);ch=children(h);wait(lambda:text(ch[109]));return h
def view():return json.loads(text(ch[109]))
def open_settings():
 click(ch[115]);s=wait(lambda:next((x for x in wins(p.pid) if text(x)=='VPN монитор — Настройки'),None),5);assert not u.IsWindowEnabled(h);return s,children(s)
def pause():
 s,ed=open_settings();before=view()['Paused'];click(ed[226]);assert text(ed[226])==('Пауза' if before else 'Продолжить');click(ed[222]);wait(lambda:view()['Paused']!=before)
def slider(s,ed,n):
 u.SendMessageW(ed[224],0x405,1,n);u.SendMessageW(s,0x114,5,ed[224]);assert text(ed[225])==str(n)+'%'
def alpha():
 a=w.BYTE();flags=w.DWORD();key=w.DWORD()
 assert u.GetLayeredWindowAttributes(h,c.byref(key),c.byref(a),c.byref(flags));assert flags.value==2
 return a.value
u.GetLayeredWindowAttributes.argtypes=[w.HWND,c.POINTER(w.DWORD),c.POINTER(w.BYTE),c.POINTER(w.DWORD)];u.GetLayeredWindowAttributes.restype=w.BOOL
def snap(win,name):
 u.SetForegroundWindow(win);time.sleep(.3);ImageGrab.grab(bbox=tuple(rect(win)),all_screens=True).save(out/name)
def close():
 click(ch[114]);p.wait(5);assert p.returncode==0
try:
 b=exe.read_bytes();pe=struct.unpack_from('<I',b,60)[0];assert struct.unpack_from('<H',b,pe+4)[0]==0x8664 and struct.unpack_from('<H',b,pe+24+68)[0]==2
 launch();defaults=view()['Config'];original=rect(h);report['dpi']=u.GetDpiForWindow(h);report['main_rect']=original
 scale=report['dpi']/96;assert original[2]-original[0]==round(380*scale) and original[3]-original[1]==round(310*scale)
 assert all(i not in ch for i in (101,102,107));assert alpha()==255 and not u.GetWindowLongPtrW(h,-20)&0x20
 report['checks']['main_smaller_old_controls_absent']=True;snap(h,'mainopaque.png')
 rs={i:rect(ch[i]) for i in (110,113,114,115)}
 for i,a in rs.items():
  for j,b in rs.items():
   if i<j:assert not(a[0]<b[2] and b[0]<a[2] and a[1]<b[3] and b[1]<a[3])
 report['checks']['gear_no_collisions']=True
 s,ed=open_settings()
 for i in range(201,212):
  style=u.GetWindowLongPtrW(ed[i],-16)
  assert not style&0x100000 and style&0x10000 and (style&0x800000 or u.GetWindowLongPtrW(ed[i],-20)&0x200),('edit style',i,hex(style),hex(u.GetWindowLongPtrW(ed[i],-20)))
 report['checks']['editable_borders_tab_no_scrollbars']=True
 assert not u.GetWindowLongPtrW(s,-20)&0x80000
 u.SendMessageW(ed[224],0x100,0x27,0);u.SendMessageW(ed[224],0x101,0x27,0);assert u.SendMessageW(ed[224],0x400,0,0)==1 and text(ed[225])=='1%';report['checks']['native_slider_keyboard']=True
 slider(s,ed,70);snap(s,'settings.png');settext(ed[201],'Cancelled');click(ed[222]);assert u.IsWindowEnabled(h) and view()['Config']==defaults and not cfg.exists() and alpha()==255;report['checks']['cancel_no_side_effect']=True
 s,ed=open_settings();settext(ed[202],'::1');click(ed[221]);assert 'Узел 1' in text(ed[220]) and view()['Config']==defaults and not cfg.exists();snap(s,'settings-invalid.png');report['checks']['invalid_ipv4_rejected']=True
 for i in range(4):settext(ed[201+i*2],'Local '+str(i+1));settext(ed[202+i*2],'127.0.0.1')
 settext(ed[209],f'127.0.0.1:{sock.getsockname()[1]}');settext(ed[210],'fixture.test');settext(ed[211],'250')
 # A directory at the exact save target proves persistence failure is not runtime success.
 slider(s,ed,40);cfg.mkdir();click(ed[221]);assert 'Не сохранено' in text(ed[220]) and view()['Config']==defaults and alpha()==255;cfg.rmdir();report['checks']['save_failure_preserves_runtime']=True
 snap(s,'settings-loopback.png');click(ed[221]);saved=json.loads(cfg.read_text(encoding='utf8'));assert saved['targets'][3]['name']=='Local 4'
 wait(lambda:all(n['Seen'] and n['OK'] for n in view()['Nodes']) and text(ch[104]).count('таймаут 250 мс')==5)
 assert 'fixture.test' in queries and all('127.0.0.1' in row for row in text(ch[104]).splitlines());report['checks']['real_loopback_four_icmp_and_custom_dns']=True;report['rows']=text(ch[104]);snap(h,'main-loopback.png')
 wait(lambda:len(view()['Samples'])>=2);report['real_samples']=view()['Samples'];assert view()['YouTubeOutages']==0
 assert alpha()==153;report['checks']['real_alpha_40']=True;snap(h,'maintransparent.png')
 click(ch[110]);assert u.GetWindowLongPtrW(h,-20)&8;click(ch[110]);assert not u.GetWindowLongPtrW(h,-20)&8
 pause();time.sleep(.5);frozen=view();time.sleep(1.5);assert frozen['Paused'] and frozen['Nodes']==view()['Nodes'] and frozen['Cycles']==view()['Cycles'];
 s,ed=open_settings();assert text(ed[226])=='Продолжить';slider(s,ed,70);click(ed[221]);wait(lambda:view()['Config']['transparency_percent']==70);assert alpha()==77 and view()['Nodes']==frozen['Nodes'] and view()['Samples']==frozen['Samples'] and view()['YouTubeOutages']==frozen['YouTubeOutages'];report['checks']['appearance_preserves_history']=True
 click(ch[110]);assert u.GetWindowLongPtrW(h,-20)&8 and alpha()==77
 s,ed=open_settings();assert not u.GetWindowLongPtrW(s,-20)&0x80000;slider(s,ed,0);click(ed[221]);assert alpha()==255 and u.GetWindowLongPtrW(h,-20)&8
 s,ed=open_settings();slider(s,ed,40);click(ed[221]);assert alpha()==153 and u.GetWindowLongPtrW(h,-20)&8;click(ch[110]);assert not u.GetWindowLongPtrW(h,-20)&8 and alpha()==153
 pause();wait(lambda:not view()['Paused']);report['checks']['pin_pause_resume_alpha_preserved']=True
 click(ch[111]);expanded=rect(h);assert expanded[:2]==original[:2] and expanded[2]==original[2] and expanded[3]-expanded[1]==round(494*scale);assert u.IsWindowVisible(ch[112]);click(ch[108]);wait(lambda:any('Проверка   ·' in x for x in view()['Journal']));click(ch[108]);snap(h,'main-journal.png');click(ch[111]);assert rect(h)==original;report['checks']['journal_verbose_compact']=True
 s,ed=open_settings();settext(ed[211],'137');click(ed[221]);assert json.loads(cfg.read_text(encoding='utf8'))['timeout_ms']==137 and view()['Config']['timeout_ms']==137
 for invalid in ['0','60001','abc','-1','1.5']:
  s,ed=open_settings();settext(ed[211],invalid);click(ed[221]);assert '1 до 60000' in text(ed[220]) and view()['Config']['timeout_ms']==137;click(ed[222])
 for boundary in ['1','60000','137']:
  s,ed=open_settings();settext(ed[211],boundary);click(ed[221]);assert view()['Config']['timeout_ms']==int(boundary)
 report['checks']['settings_timeout_validation_boundaries']=True
 for win in wins(p.pid):
  cls=c.create_unicode_buffer(100);u.GetClassNameW(win,cls,100);assert cls.value!='ConsoleWindowClass'
 report['checks']['no_console_amd64_gui']=True
 close();launch();assert view()['Config']['timeout_ms']==137 and view()['Config']['targets'][3]['name']=='Local 4' and alpha()==153;wait(lambda:all(n['Seen'] and n['OK'] for n in view()['Nodes']));report['checks']['restart_persistence']=True
 # Change fourth identity while paused: old samples/counter must clear, no synthetic probes.
 pause();time.sleep(.5);s,ed=open_settings();settext(ed[207],'Fourth renamed');click(ed[221]);assert not view()['Samples'] and view()['YouTubeOutages']==0 and not view()['Nodes'][3]['Seen'];report['checks']['fourth_identity_resets']=True
 s,ed=open_settings();click(ed[223]);click(ed[222]);assert view()['Config']['targets'][3]['name']=='Fourth renamed';report['checks']['defaults_cancel_not_applied']=True
 s,ed=open_settings();click(ed[223]);assert u.SendMessageW(ed[224],0x400,0,0)==0 and text(ed[211])=='2000';click(ed[221]);assert view()['Config']==defaults and alpha()==255;close()
 old=dict(defaults);del old['transparency_percent'];cfg.write_text(json.dumps(old),encoding='utf8');launch();assert view()['Config']==defaults and alpha()==255;close();report['checks']['old_version1_missing_transparency']=True
 for bad in [-1,71,'bad']:
  invalid=dict(defaults);invalid['transparency_percent']=bad;raw=json.dumps(invalid);cfg.write_text(raw,encoding='utf8');launch();assert view()['Config']==defaults and alpha()==255 and cfg.read_text(encoding='utf8')==raw;close()
 report['checks']['invalid_transparency_preserved']=True
 # Corrupt original stays byte-identical until an explicit successful Save.
 cfg.write_text('{malformed',encoding='utf8');launch();assert view()['Config']==defaults and cfg.read_text()=='{malformed';assert any('Не удалось загрузить' in x for x in view()['Journal']);s,ed=open_settings();assert 'Не удалось загрузить' in text(ed[220]);click(ed[222]);assert cfg.read_text()=='{malformed';report['checks']['malformed_visible_preserved']=True
 s,ed=open_settings();click(ed[223]);click(ed[221]);assert json.loads(cfg.read_text(encoding='utf8'))==defaults;close();report['checks']['defaults_restore_and_close_zero']=True
 report['dns_queries']=queries;report['passed']=True
 (out/'settings-verification.json').write_text(json.dumps(report,ensure_ascii=False,indent=2),encoding='utf8');print(json.dumps(report,ensure_ascii=False,indent=2))
finally:
 if p is not None and p.poll() is None:p.kill();p.wait()
 stop.set();sock.close()
 if cfg.is_file():cfg.unlink()

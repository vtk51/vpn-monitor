"""Original shield/network pulse mark; optional build-time dependency: Pillow.
Run python generate_icon.py, then regenerate-resources.cmd.
"""
from PIL import Image, ImageDraw, ImageFont
from pathlib import Path
import struct

ROOT=Path(__file__).resolve().parent
SIZES=(16,20,24,32,40,48,64,128,256)

def icon(n):
    scale=4
    im=Image.new('RGBA',(n*scale,n*scale))
    d=ImageDraw.Draw(im)
    def pts(p): return [(round(x*n*scale/256),round(y*n*scale/256)) for x,y in p]
    def line(p,color,width):d.line(pts(p),fill=color,width=max(1,round(width*n*scale/256)),joint='curve')
    d.rounded_rectangle((2*scale*n/256,2*scale*n/256,254*scale*n/256,254*scale*n/256),radius=54*n*scale/256,fill='#171e28',outline='#526073',width=max(1,round(3*n*scale/256)))
    # Quiet graphite highlight and the distinctive open shield silhouette.
    d.rounded_rectangle((10*n*scale/256,10*n*scale/256,246*n*scale/256,246*n*scale/256),radius=46*n*scale/256,outline='#283442',width=max(1,round(2*n*scale/256)))
    shield=[(128,40),(198,68),(192,133),(179,165),(157,188),(128,207),(99,188),(77,165),(64,133),(58,68),(128,40)]
    d.polygon(pts(shield),fill='#152d47')
    line(shield,'#56aaff',10 if n>=24 else 13)
    if n>=32:
        line([(91,94),(128,77),(165,94)],'#387abd',4)
        line([(128,77),(128,114)],'#387abd',4)
        for x,y in [(91,94),(128,77),(165,94)]:
            r=4
            d.ellipse((*pts([(x-r,y-r)])[0],*pts([(x+r,y+r)])[0]),fill='#8ac6ff')
    pulse=[(41,136),(91,136),(110,109),(131,161),(150,126),(163,136),(215,136)]
    line(pulse,'#102b2c',20 if n>=24 else 23)
    line(pulse,'#57efa8',10 if n>=24 else 13)
    return im.resize((n,n),Image.Resampling.LANCZOS)

frames=[icon(n) for n in SIZES]
# Explicit uncompressed 32-bit DIB frames, compatible with legacy resource tools.
parts=[]
for im in frames:
    n=im.width
    bgra=im.transpose(Image.Transpose.FLIP_TOP_BOTTOM).tobytes('raw','BGRA')
    stride=((n+31)//32)*4
    mask=bytes(stride*n)
    parts.append(struct.pack('<IiiHHIIiiII',40,n,2*n,1,32,0,len(bgra)+len(mask),0,0,0,0)+bgra+mask)
offset=6+16*len(parts)
entries=[]
for n,data in zip(SIZES,parts):
    entries.append(struct.pack('<BBBBHHII',n%256,n%256,0,0,1,32,len(data),offset));offset+=len(data)
(ROOT/'app.ico').write_bytes(struct.pack('<HHH',0,1,len(parts))+b''.join(entries)+b''.join(parts))
preview=Image.new('RGB',(820,400),'#10151d');d=ImageDraw.Draw(preview)
d.text((28,22),'VPN MONITOR / SHIELD + NETWORK + LIVE PULSE',fill='#a9bed5')
large=icon(256);preview.paste(large,(34,88),large)
for x,n in [(350,16),(440,32),(545,48),(665,64)]:
    im=icon(n);preview.paste(im,(x+(64-n)//2,160+(64-n)//2),im)
    d.text((x+14,244),f'{n}px',fill='#b8c8d9')
d.text((350,310),'Native sizes / no scaling',fill='#6f8ba7')
preview.save(ROOT/'icon-preview.png')
print('Generated app.ico:',SIZES)

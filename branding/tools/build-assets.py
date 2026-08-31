#!/usr/bin/env python3
from pathlib import Path
from PIL import Image, ImageDraw, ImageFont
import os, shutil, subprocess, sys

ROOT = Path(__file__).resolve().parents[1]
FONT_DIR = Path(os.environ.get('AKILIX_FONT_DIR', '/usr/share/fonts/truetype/dejavu'))
FULL = Image.open(ROOT/'source/akilix-master.png').convert('RGBA')
HEAD = Image.open(ROOT/'source/akilix-mark-master.png').convert('RGBA')
BG = '#0B1114'; GRAPHITE = '#151A1D'; BONE = '#E8E6DD'; GREEN = '#657A3E'; LEAF = '#8EAD55'; AMBER = '#C68A2B'; SLATE = '#4A5358'

def font(size, bold=False):
    p = FONT_DIR / ('DejaVuSans-Bold.ttf' if bold else 'DejaVuSans.ttf')
    return ImageFont.truetype(p, size)
def contain(im, box):
    x, y, w, h = box; copy = im.copy(); copy.thumbnail((w, h), Image.Resampling.LANCZOS)
    return copy, (x+(w-copy.width)//2, y+(h-copy.height)//2)
def place(bg, im, box): bg.alpha_composite(*contain(im, box))
def text_logo(size, dark=True, tagline=True):
    im = Image.new('RGBA', size, (0,0,0,0)); d=ImageDraw.Draw(im); f=font(max(24,size[1]//4), True)
    x=size[0]//3; y=size[1]//3
    d.text((x,y), 'Aki', font=f, fill=BONE if dark else GRAPHITE)
    prefixw=d.textbbox((x,y),'Aki',font=f)[2]-x
    d.text((x+prefixw-3,y), 'lix', font=f, fill=LEAF if dark else GREEN)
    if tagline: d.text((x+2,y+f.size+10), 'SECURITY WORK WITH PROVENANCE.', font=font(max(12,size[1]//25)), fill='#AAB1AF' if dark else SLATE)
    return im
def logo(size, dark=True, tagline=True):
    out=Image.new('RGBA',size,(0,0,0,0)); place(out,FULL,(0,0,int(size[0]*.36),size[1])); out.alpha_composite(text_logo(size,dark,tagline)); return out
def save(im, rel, size=None, bg=None):
    if size: im=im.resize(size,Image.Resampling.LANCZOS)
    if bg:
        base=Image.new('RGBA',im.size,bg); base.alpha_composite(im); im=base.convert('RGB')
    im.save(ROOT/rel,'PNG',optimize=False)
def ensure():
    for p in ['web','desktop','os/plymouth','os/grub','os/installer','os/wallpaper','print']: (ROOT/p).mkdir(parents=True,exist_ok=True)
def main():
    ensure(); darklogo=logo((1200,420),True); lightlogo=logo((1200,420),False)
    save(darklogo,'web/akilix-horizontal.png',(1200,420),BG); save(darklogo,'web/akilix-horizontal@2x.png',(2400,840),BG)
    save(darklogo,'web/akilix-horizontal-dark.png',(1200,420),BG); save(lightlogo,'web/akilix-horizontal-light.png',(1200,420),'#E8E6DD')
    save(HEAD,'web/akilix-mark.png',(900,900)); save(HEAD,'desktop/akilix.svg') if False else None
    for s in [16,22,24,32,48,64,128,256,512]: save(HEAD,f'desktop/akilix-{s}.png',(s,s))
    save(HEAD,'web/favicon-16.png',(16,16)); save(HEAD,'web/favicon-32.png',(32,32)); save(HEAD,'web/favicon-48.png',(48,48)); save(HEAD,'web/apple-touch-icon.png',(180,180)); save(HEAD,'web/android-chrome-192.png',(192,192)); save(HEAD,'web/android-chrome-512.png',(512,512))
    save(darklogo,'web/github-social-1280x640.png',(1280,640),BG); save(darklogo,'web/social-card-1200x630.png',(1200,630),BG)
    magick = shutil.which('magick')
    if magick:
        subprocess.run([magick,str(ROOT/'web/favicon-16.png'),str(ROOT/'web/favicon-32.png'),str(ROOT/'web/favicon-48.png'),str(ROOT/'web/favicon.ico')],check=True)
    elif not (ROOT/'web/favicon.ico').is_file():
        raise RuntimeError('ImageMagick is required to create web/favicon.ico')
    ply=Image.new('RGBA',(1920,1080),BG); place(ply,HEAD,(700,120,520,520)); d=ImageDraw.Draw(ply); d.text((750,700),'Akilix',font=font(74,True),fill=BONE); d.text((753,790),'Security work with provenance.',font=font(24),fill=SLATE); save(ply,'os/plymouth/splash-1920x1080.png')
    save(HEAD,'os/plymouth/logo.png',(512,512)); save(HEAD,'os/grub/logo.png',(420,420))
    grub=Image.new('RGBA',(1920,1080),BG); place(grub,FULL,(1180,180,650,700)); save(grub,'os/grub/background-1920x1080.png')
    save(darklogo,'os/installer/logo.png',(800,280),BG); save(darklogo,'os/installer/banner.png',(1400,360),BG)
    for s in [(1920,1080),(2560,1440),(3840,2160)]:
        wall=Image.new('RGBA',s,BG); place(wall,FULL,(int(s[0]*.53),int(s[1]*.35),int(s[0]*.44),int(s[1]*.55))); save(wall,f'os/wallpaper/akilix-{s[0]}x{s[1]}.png')
    sticker=Image.new('RGBA',(1200,1200),'#E8E6DD'); place(sticker,FULL,(40,40,720,820)); d=ImageDraw.Draw(sticker); d.text((750,480),'Akilix',font=font(72,True),fill=GRAPHITE); d.text((755,570),'Security work with provenance.',font=font(20),fill=SLATE); save(sticker,'print/akilix-sticker.png')
    contact=Image.new('RGB',(1600,1380),'#20282B'); cd=ImageDraw.Draw(contact); cd.text((40,25),'Akilix BRANDING REVIEW',font=font(32,True),fill=BONE)
    def terminal_card(lines):
        card=Image.new('RGBA',(720,210),BG); td=ImageDraw.Draw(card)
        for n,line in enumerate(lines): td.text((20,20+n*25),line,font=font(20),fill=BONE)
        return card
    cards=[('PRIMARY',darklogo),('COMPACT',HEAD),('16px × 12',HEAD.resize((192,192),Image.Resampling.NEAREST)),('32px × 12',HEAD.resize((384,384),Image.Resampling.NEAREST)),('SOCIAL',Image.open(ROOT/'web/social-card-1200x630.png')),('PLYMOUTH',ply),('GRUB',grub),('WALLPAPER',Image.open(ROOT/'os/wallpaper/akilix-1920x1080.png')),('TERMINAL FULL',terminal_card(['                         __','                    _.-\'  `-._','              _..--\'  _      `-._','       ______/      _/ \\_         `\\','Akilix  Security work with provenance.'])),('TERMINAL SMALL',terminal_card(['      __',' _..-\'  `-._','/___  ___  _\\','   `-o-o-o-\'  Akilix']))]
    for i,(label,im) in enumerate(cards):
        x=40+(i%2)*780; y=85+(i//2)*255; cd.text((x,y),label,font=font(18,True),fill=LEAF); thumb=im.convert('RGBA'); thumb.thumbnail((720,210)); contact.paste(thumb,(x,y+28),thumb)
    contact.save(ROOT/'preview.png','PNG',optimize=False)

    # Reference boards are generated from canonical Akilix assets so historical
    # exploratory artwork cannot accidentally reintroduce the retired name.
    board1=Image.new('RGBA',(1536,1024),GRAPHITE); bd=ImageDraw.Draw(board1)
    bd.text((64,48),'AKILIX IDENTITY',font=font(44,True),fill=BONE)
    bd.text((67,108),'Security work with provenance.',font=font(22),fill=SLATE)
    place(board1,FULL,(40,180,760,760)); place(board1,darklogo,(720,250,760,360))
    palette=[BG,GRAPHITE,BONE,GREEN,LEAF,AMBER,SLATE]
    for i,color in enumerate(palette): bd.rectangle((760+i*92,720,832+i*92,792),fill=color)
    board1.convert('RGB').save(ROOT/'reference/concept-board-01.png','PNG',optimize=False)

    board2=Image.new('RGBA',(1536,1024),BG); bd=ImageDraw.Draw(board2)
    bd.text((64,48),'AKILIX SYSTEM SURFACES',font=font(40,True),fill=BONE)
    samples=[('COMPACT MARK',HEAD),('BOOT',ply),('DESKTOP',Image.open(ROOT/'os/wallpaper/akilix-1920x1080.png')),('SOCIAL',Image.open(ROOT/'web/social-card-1200x630.png'))]
    for i,(label,sample) in enumerate(samples):
        x=64+(i%2)*736; y=130+(i//2)*420
        bd.text((x,y),label,font=font(18,True),fill=LEAF)
        thumb=sample.convert('RGBA'); thumb.thumbnail((680,360)); board2.alpha_composite(thumb,(x,y+38))
    board2.convert('RGB').save(ROOT/'reference/concept-board-02.png','PNG',optimize=False)
if __name__=='__main__': main()

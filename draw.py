from matplotlib import pyplot as plt 
import numpy as np 

f = open("result_for_draw.txt",encoding = "utf-8")

xs=[]
mynsop=[]
myBop=[]
nsop=[]
Bop=[]
while 1:
    line = f.readline()
    if line is not None and line != '':
        xs.append(int(line.split('=')[1]))

        line = f.readline()
        line = ' '.join(line.split()) 
        tmp = line.split(' ')
        mynsop.append(float(tmp[2]))
        myBop.append(float(tmp[4]))

        line = f.readline()
        line = ' '.join(line.split()) 
        tmp = line.split(' ')
        nsop.append(float(tmp[2]))
        Bop.append(float(tmp[4]))
    else:
        break 

print("xs : ",xs)
print("nsop : ",nsop)
print("Bop : ",Bop)

plt.figure(dpi=500)
plt.subplot(211)
plt.title("Benchmark Results : Prependable Buffer && []byte ") 
# plt.xlabel("read data len") 
plt.ylabel("ns/op") 
plt.plot(xs,mynsop,label="Prependable buffer")
plt.plot(xs,nsop,label="normal buffer")
plt.legend(loc="upper left")

plt.subplot(212)
plt.xlabel("read data len") 
plt.ylabel("B/op") 
plt.plot(xs,myBop,label="Prependable buffer")
plt.plot(xs,Bop,label="normal buffer")
plt.legend(loc="upper left")

plt.savefig('./report.jpg')
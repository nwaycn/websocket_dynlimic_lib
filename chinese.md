## 说明   
    这些年随着做智能外呼、智能客服比较多，公网中的讯飞、百度、阿里、腾讯的asr/tts，私网中的基于mrcp协议的各asr/tts厂商，以及各式各样的私有的sdk,webapi,websocket等等（厂商可能就比较多了，nuance、讯飞、百度、思必驰、云知声、阿里、捷通华声、谷歌、腾讯、标贝、以及各种基于tensorflow/torch/kaldi等开源训练产品生成的asr/tts),此文是基于采用websocket传输数据，之前都是用的c语言进行开发，后来感觉c在有些方面不方便，所以就用go来实现。
## 使用
    
### 1. go version

go version go1.16.4 linux/amd64

### 2. gcc --version

gcc (GCC) 4.4.7 20120313 (Red Hat 4.4.7-23) Copyright (C) 2010 Free Software Foundation, Inc. This is free software; see the source for copying conditions. There is NO warranty; not even for MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.

### 3. 编译成动态链接库

./build.sh

### 4. 在c语言代码中调用

//包含头文件 
#include "libnway_websocket.h"
//初始化库
if (nway_asr_init()!=0){
   goto END;
}

//创建新连接
char* errmsg=malloc(100);
nway_memset(rh->r_nway_sid,globals.nway_sid_len);
recStatus = nway_asr_connect(globals.nway_uri,rh->r_nway_sid,errmsg,"","");
if (0 != recStatus)
{ 
}else{
  nway_safe_free(errmsg);
}

 //发送数据调用
char* message=malloc(globals.nway_message_len);
nway_memset(message,globals.nway_message_len);
recStatus = nway_asr_sendmessage(rh->r_nway_sid,(void*)firstData,firstSamples,message);
if(recStatus == 0)
{
	switch_log_printf(SWITCH_CHANNEL_SESSION_LOG(rh->session), SWITCH_LOG_DEBUG, "sent data\n");

}else{
	switch_log_printf(SWITCH_CHANNEL_SESSION_LOG(rh->session), SWITCH_LOG_ERROR, "sbb error:%s \n",message);

}

//当有vad即一句话说完时

char* message=malloc(globals.nway_message_len);;

nway_asr_stop(rh->r_nway_sid,message);
        
### 5. 编译c代码

gcc -o test test.c -L./ -lnway_websocket

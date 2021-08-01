# websocket_dynlimic_lib
How to call a simple websocket dynlimic library with go ?

1. go version
   go version go1.16.4 linux/amd64
   
2. gcc --version
   gcc (GCC) 4.4.7 20120313 (Red Hat 4.4.7-23)
Copyright (C) 2010 Free Software Foundation, Inc.
This is free software; see the source for copying conditions.  There is NO
warranty; not even for MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.

3. build go code to dynlimic library
   ./build.sh
   
4. call lib with c
   #include "libnway_websocket.h"
   //init library
   if (nway_asr_init()!=0){
      goto END;
    }
    
    //call function
        char* message=malloc(globals.nway_message_len);
				nway_memset(message,globals.nway_message_len);
				recStatus = nway_asr_sendmessage(rh->r_nway_sid,(void*)firstData,firstSamples,message);
				if(recStatus == 0)
				{
					switch_log_printf(SWITCH_CHANNEL_SESSION_LOG(rh->session), SWITCH_LOG_DEBUG, "sent data\n");
						
				}else{
					switch_log_printf(SWITCH_CHANNEL_SESSION_LOG(rh->session), SWITCH_LOG_ERROR, "sbb error:%s \n",message);
					 
				}
        
 5. compile c code
     gcc -o test test.c -L./ -lnway_websocket

#include "libnway_websocket.h"
#include <stdlib.h>
#include <stdio.h>

int main(){
    char id[200];
    get_uuid(id,200);
    printf("get id in c:%s\n",id);
    return 0;
}

#include <stdio.h>
#include "lists.h"
extern void load_seed(char *);
int main(void) {
    load_seed("seed.bin");
    ilist l = NULL;
    for (int i = 0; i < 50; i++) ilist_append(&l, i);
    ilist_shuffle(l);
    int n = ilist_len(l);
    for (int i = 0; i < n; i++) printf("%d\n", l[i]);
    return 0;
}

/* LibTomCrypt, modular cryptographic library -- Tom St Denis
 *
 * LibTomCrypt is a library that provides various cryptographic
 * algorithms in a highly modular and flexible manner.
 *
 * The library is free for all purposes without any express
 * guarantee it works.
 */
#include "tomcrypt_private.h"

/**
   @file cbc_getiv.c
   CBC implementation, get IV, Tom St Denis
*/

#ifdef LTC_CBC_MODE

/**
   Get the current initialization vector
   @param IV   [out] The destination of the initialization vector
   @param len  [in/out]  The max size and resulting size of the initialization vector
   @param cbc  The CBC state
   @return CRYPT_OK if successful
*/
int cbc_getiv(unsigned char *IV, unsigned long *len, const symmetric_CBC *cbc)
{
   LTC_ARGCHK(IV  != NULL);
   LTC_ARGCHK(len != NULL);
   LTC_ARGCHK(cbc != NULL);
   if ((unsigned long)cbc->blocklen > *len) {
      *len = cbc->blocklen;
      return CRYPT_BUFFER_OVERFLOW;
   }
   XMEMCPY(IV, cbc->IV, cbc->blocklen);
   *len = cbc->blocklen;

   return CRYPT_OK;
}

#endif

/* ref:         $Format:%D$ */
/* git commit:  $Format:%H$ */
/* commit time: $Format:%ai$ */

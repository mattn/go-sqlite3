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
   @file cbc_setiv.c
   CBC implementation, set IV, Tom St Denis
*/


#ifdef LTC_CBC_MODE

/**
   Set an initialization vector
   @param IV   The initialization vector
   @param len  The length of the vector (in octets)
   @param cbc  The CBC state
   @return CRYPT_OK if successful
*/
int cbc_setiv(const unsigned char *IV, unsigned long len, symmetric_CBC *cbc)
{
   LTC_ARGCHK(IV  != NULL);
   LTC_ARGCHK(cbc != NULL);
   if (len != (unsigned long)cbc->blocklen) {
      return CRYPT_INVALID_ARG;
   }
   XMEMCPY(cbc->IV, IV, len);
   return CRYPT_OK;
}

#endif


/* ref:         $Format:%D$ */
/* git commit:  $Format:%H$ */
/* commit time: $Format:%ai$ */

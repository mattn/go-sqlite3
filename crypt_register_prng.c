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
  @file crypt_register_prng.c
  Register a PRNG, Tom St Denis
*/

/**
   Register a PRNG with the descriptor table
   @param prng   The PRNG you wish to register
   @return value >= 0 if successfully added (or already present), -1 if unsuccessful
*/
int register_prng(const struct ltc_prng_descriptor *prng)
{
   int x;

   LTC_ARGCHK(prng != NULL);

   /* is it already registered? */
   LTC_MUTEX_LOCK(&ltc_prng_mutex);
   for (x = 0; x < TAB_SIZE; x++) {
       if (XMEMCMP(&prng_descriptor[x], prng, sizeof(struct ltc_prng_descriptor)) == 0) {
          LTC_MUTEX_UNLOCK(&ltc_prng_mutex);
          return x;
       }
   }

   /* find a blank spot */
   for (x = 0; x < TAB_SIZE; x++) {
       if (prng_descriptor[x].name == NULL) {
          XMEMCPY(&prng_descriptor[x], prng, sizeof(struct ltc_prng_descriptor));
          LTC_MUTEX_UNLOCK(&ltc_prng_mutex);
          return x;
       }
   }

   /* no spot */
   LTC_MUTEX_UNLOCK(&ltc_prng_mutex);
   return -1;
}

/* ref:         $Format:%D$ */
/* git commit:  $Format:%H$ */
/* commit time: $Format:%ai$ */

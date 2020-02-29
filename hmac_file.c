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
  @file hmac_file.c
  HMAC support, process a file, Tom St Denis/Dobes Vandermeer
*/

#ifdef LTC_HMAC

/**
  HMAC a file
  @param hash     The index of the hash you wish to use
  @param fname    The name of the file you wish to HMAC
  @param key      The secret key
  @param keylen   The length of the secret key
  @param out      [out] The HMAC authentication tag
  @param outlen   [in/out]  The max size and resulting size of the authentication tag
  @return CRYPT_OK if successful, CRYPT_NOP if file support has been disabled
*/
int hmac_file(int hash, const char *fname,
              const unsigned char *key, unsigned long keylen,
                    unsigned char *out, unsigned long *outlen)
{
#ifdef LTC_NO_FILE
   LTC_UNUSED_PARAM(hash);
   LTC_UNUSED_PARAM(fname);
   LTC_UNUSED_PARAM(key);
   LTC_UNUSED_PARAM(keylen);
   LTC_UNUSED_PARAM(out);
   LTC_UNUSED_PARAM(outlen);
    return CRYPT_NOP;
#else
   hmac_state hmac;
   FILE *in;
   unsigned char *buf;
   size_t x;
   int err;

   LTC_ARGCHK(fname  != NULL);
   LTC_ARGCHK(key    != NULL);
   LTC_ARGCHK(out    != NULL);
   LTC_ARGCHK(outlen != NULL);

   if ((buf = XMALLOC(LTC_FILE_READ_BUFSIZE)) == NULL) {
      return CRYPT_MEM;
   }

   if ((err = hash_is_valid(hash)) != CRYPT_OK) {
      goto LBL_ERR;
   }

   if ((err = hmac_init(&hmac, hash, key, keylen)) != CRYPT_OK) {
      goto LBL_ERR;
   }

   in = fopen(fname, "rb");
   if (in == NULL) {
      err = CRYPT_FILE_NOTFOUND;
      goto LBL_ERR;
   }

   do {
      x = fread(buf, 1, LTC_FILE_READ_BUFSIZE, in);
      if ((err = hmac_process(&hmac, buf, (unsigned long)x)) != CRYPT_OK) {
         fclose(in); /* we don't trap this error since we're already returning an error! */
         goto LBL_CLEANBUF;
      }
   } while (x == LTC_FILE_READ_BUFSIZE);

   if (fclose(in) != 0) {
      err = CRYPT_ERROR;
      goto LBL_CLEANBUF;
   }

   err = hmac_done(&hmac, out, outlen);

LBL_CLEANBUF:
   zeromem(buf, LTC_FILE_READ_BUFSIZE);
LBL_ERR:
#ifdef LTC_CLEAN_STACK
   zeromem(&hmac, sizeof(hmac_state));
#endif
   XFREE(buf);
   return err;
#endif
}

#endif

/* ref:         $Format:%D$ */
/* git commit:  $Format:%H$ */
/* commit time: $Format:%ai$ */

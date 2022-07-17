using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Security.Cryptography;
using System.Text;
using System.Threading.Tasks;
// 참고사이트 : https://gods2000.tistory.com/entry/AES256-%EB%B3%B5%ED%98%B8%EC%95%94%ED%98%B8%ED%99%94-%ED%95%98%EB%8A%94-%EB%B0%A9%EB%B2%95C
namespace NSPCCSIS
{
    public static class AES
    {
        public static string Encrypt256(string Input, string key, string iv)
        {
            RijndaelManaged aes = new RijndaelManaged();
            aes.KeySize = 256;
            aes.BlockSize = 128;
            aes.Mode = CipherMode.CBC;
            // Java에서는 PKCS5 패딩을 쓰고 C#에서는 PKCS7 패딩을 쓴다. 결과적으로 상호 호환은 되나, 차이가 있다는건 알고있는게 좋을듯..
            aes.Padding = PaddingMode.PKCS7;
            aes.Key = Encoding.UTF8.GetBytes(key);
            aes.IV = Encoding.UTF8.GetBytes(iv);
            //IV값을 아래처럼 설정할수 있으나, byte값을 직접 넣는 일이 필요할지는 모르겠다.
            //aes.IV = new byte[] { 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 };

            var encrypt = aes.CreateEncryptor(aes.Key, aes.IV);
            byte[] xBuff = null;

            using (var ms = new MemoryStream())
            {
                using (var cs = new CryptoStream(ms, encrypt, CryptoStreamMode.Write))
                {
                    byte[] xXml = Encoding.UTF8.GetBytes(Input);
                    cs.Write(xXml, 0, xXml.Length);
                }
                xBuff = ms.ToArray();
            }

            string Output = Convert.ToBase64String(xBuff);

            return Output;
        }
        public static string Decrypt256(string Input, string key, string iv)
        {
            RijndaelManaged aes = new RijndaelManaged();
            aes.KeySize = 256;
            aes.BlockSize = 128;
            aes.Mode = CipherMode.CBC;
            // Java에서는 PKCS5 패딩을 쓰고 C#에서는 PKCS7 패딩을 쓴다. 결과적으로 상호 호환은 되나, 차이가 있다는건 알고있는게 좋을듯..
            aes.Padding = PaddingMode.PKCS7;
            aes.Key = Encoding.UTF8.GetBytes(key);
            aes.IV = Encoding.UTF8.GetBytes(iv);
            //IV값을 아래처럼 설정할수 있으나, byte값을 직접 넣는 일이 필요할지는 모르겠다.
            //aes.IV = new byte[] { 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 };
            
            var decrypt = aes.CreateDecryptor();
            byte[] xBuff = null;

            using (var ms = new MemoryStream())
            {
                using (var cs = new CryptoStream(ms, decrypt, CryptoStreamMode.Write))
                {
                    byte[] xXml = Convert.FromBase64String(Input);
                    cs.Write(xXml, 0, xXml.Length);
                }
                xBuff = ms.ToArray();
            }

            string Output = Encoding.UTF8.GetString(xBuff);

            return Output;
        }
    }
}

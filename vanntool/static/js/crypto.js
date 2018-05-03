
//function AESEncrypto(str, key){
//	var parsekey = CryptoJS.enc.ASCII.parse(key);
//	var pkey = PaddingLeft(parsekey, 16);
//	var encrypted = CryptoJS.AES.encrypt(str, pkey, {
//		iv: pkey,
//		mode: CryptoJS.mode.CBC,
//		padding: CryptoJS.pad.Pkcs7
//	});
//	var encryptedStr = encrypted.ciphertext.toString(CryptoJS.enc.Hex);
//	var encryptedHexStr = CryptoJS.enc.Hex.parse(encryptedStr);
//	console.log("str:",str,",pkey:",pkey, ",hexstr:", encryptedStr, ",hexparse:",encryptedHexStr);
//	return encrypted.ciphertext.toString(CryptoJS.enc.Base64);
//}

function AESEncrypto(str, key){
	if(key.length == 0){
		return "";	
	}
	key = PaddingLeft(key, 16);
	key = CryptoJS.enc.Utf8.parse(key);

	var encrypted = CryptoJS.AES.encrypt(str, key, {
		iv: key,
		mode: CryptoJS.mode.CBC,
		padding: CryptoJS.pad.Pkcs7
	});
	return  encrypted.ciphertext.toString(CryptoJS.enc.Hex);
}

function AESDecrypto(encrypted, key){
	key = PaddingLeft(key, 16);
	key = CryptoJS.enc.Utf8.parse(key);
	var decrypted = CryptoJS.AES.decrypt(encrypted, key, {
		iv: key,
		mode: CryptoJS.mode.CBC,
		padding: CryptoJS.pad.Pkcs7
	});
	 return CryptoJS.enc.Utf8.stringify(decrypted);
}

function PaddingLeft(key, length){
	pkey= key.toString();
	var l = pkey.length;
	if (l < length) {
		pkey = new Array(length - l + 1).join('0') + pkey;
	}else if (l > length){
		pkey = pkey.slice(length);
	}
	return pkey;
}


function stringToByte(str) {  
	var bytes = new Array();  
	var len, c;  
	len = str.length;  
	for(var i = 0; i < len; i++) {  
		c = str.charCodeAt(i);  
		if(c >= 0x010000 && c <= 0x10FFFF) {  
			bytes.push(((c >> 18) & 0x07) | 0xF0);  
			bytes.push(((c >> 12) & 0x3F) | 0x80);  
			bytes.push(((c >> 6) & 0x3F) | 0x80);  
			bytes.push((c & 0x3F) | 0x80);  
		} else if(c >= 0x000800 && c <= 0x00FFFF) {  
			bytes.push(((c >> 12) & 0x0F) | 0xE0);  
			bytes.push(((c >> 6) & 0x3F) | 0x80);  
			bytes.push((c & 0x3F) | 0x80);  
		} else if(c >= 0x000080 && c <= 0x0007FF) {  
			bytes.push(((c >> 6) & 0x1F) | 0xC0);  
			bytes.push((c & 0x3F) | 0x80);  
		} else {  
			bytes.push(c & 0xFF);  
		}  
	}  
	return bytes;  
}  

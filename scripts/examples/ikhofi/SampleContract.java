package ikhofi_sample;

import java.util.LinkedHashMap;
import java.util.Map;

import io.annchain.ikhofi.api.Context;
import io.annchain.ikhofi.api.TxContext;

public class SampleContract {
	Map<String, String> kvData;

	public SampleContract() {
		kvData = new LinkedHashMap<String, String>();
	}

	public void set(TxContext txContext, String key, String value) {
		if (key == null) {
			key = "";
		}
		if (value == null) {
			value = "";
		}
		txContext.put(key.getBytes(), value.getBytes());
	}

	public String get(Context ctx, String key) {
		if (key == null) {
			key = "";
		}
		return new String(ctx.get(key.getBytes()));
	}
}
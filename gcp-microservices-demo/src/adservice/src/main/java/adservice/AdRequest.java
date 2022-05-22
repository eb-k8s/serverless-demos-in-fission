package adservice;

import java.util.Arrays;
import java.util.List;

public class AdRequest {
	
	private String[] context_keys;
	
	AdRequest(String[] keys){
		this.context_keys = keys;
	}
	
	AdRequest(){
		
	}
	
	AdRequest(List<String> keys){
		this.context_keys = keys.toArray(new String[] {});
	}
	
	public void setContext_keys(String[] keys) {
		this.context_keys = keys;
	}
	
	public String[] getContext_keys() {
		return this.context_keys;
	}
	
	public int getContextKeysCount() {
		return this.context_keys.length;
	}
	
	public String getContextKeysList() {
		return Arrays.toString(this.context_keys);
	}
}

package adservice;

import java.util.List;

public class AdResponse {
	
	private Ad[] ads;
	
	AdResponse(Ad[] ads){
		this.ads = ads;
	}
	
	AdResponse(){
		
	}
	
	AdResponse(List<Ad> ads){
		this.ads = ads.toArray(new Ad[] {});
	}
	
	public void setAds(Ad[] ads) {
		this.ads = ads;
	}
	
	public Ad[] getAds() {
		return this.ads;
	}
}

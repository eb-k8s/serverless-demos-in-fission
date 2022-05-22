package adservice;

public class Ad {
	
	private String redirect_url;
	
	private String text;
	
	Ad(String url, String text){
		this.redirect_url = url;
		this.text = text;
	}
	
	Ad(){
		
	}
	
	public void setRedirect_url(String url) {
		this.redirect_url = url;
	}
	
	public String getRedirect_url() {
		return this.redirect_url;
	}
	
	public void setText(String text) {
		this.text = text;
	}
	
	public String getText() {
		return this.text;
	}
}

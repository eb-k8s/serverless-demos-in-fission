package adservice;

import static org.junit.Assert.*;

import org.junit.Test;

import java.net.URI;
import java.net.URISyntaxException;

import org.springframework.http.RequestEntity;
import org.springframework.http.ResponseEntity;
import org.springframework.util.Assert;

public class TestAdservice {

	@Test
	public void testonekey() throws URISyntaxException {
		Adservice ad = new Adservice();
		String uri = "http://example.com/ad?context_key=clothing";
		RequestEntity request = RequestEntity.get(new URI(uri)).build();
		ResponseEntity resp = ad.call(request, null);
		String body = resp.getBody().toString();
		System.out.println(body);
		Assert.hasText(body,"has body");
	}

	@Test
	public void testemptykey() throws URISyntaxException {
		Adservice ad = new Adservice();
		String uri = "http://example.com/ad";
		RequestEntity request = RequestEntity.get(new URI(uri)).build();
		ResponseEntity resp = ad.call(request, null);
		String body = resp.getBody().toString();
		System.out.println(body);
		Assert.hasText(body,"has body");
	}
	
	@Test
	public void testmorekey() throws URISyntaxException {
		Adservice ad = new Adservice();
		String uri = "http://example.com/ad?context_key=footwear&context_key=decor";
		RequestEntity request = RequestEntity.get(new URI(uri)).build();
		ResponseEntity resp = ad.call(request, null);
		String body = resp.getBody().toString();
		System.out.println(body);
		Assert.hasText(body,"has body");
	}
	
	@Test
	public void testwrongkey() throws URISyntaxException {
		Adservice ad = new Adservice();
		String uri = "http://example.com/ad?wrong_key=decor";
		RequestEntity request = RequestEntity.get(new URI(uri)).build();
		ResponseEntity resp = ad.call(request, null);
		Object rawbody = resp.getBody();
		Assert.isNull(rawbody, "the body must be null");
	}
}

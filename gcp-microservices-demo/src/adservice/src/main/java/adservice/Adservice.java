package adservice;

import java.io.UnsupportedEncodingException;
import java.lang.Exception;
import java.net.URI;
import java.net.URLDecoder;
import java.util.ArrayList;
import java.util.Collection;
import java.util.Collections;
import java.util.List;
import java.util.Random;
import java.util.logging.Level;
import java.util.logging.Logger;

import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.RequestEntity;
import org.springframework.http.ResponseEntity;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.common.collect.ImmutableListMultimap;
import com.google.common.collect.Iterables;

import io.fission.Function;
import io.fission.Context;

public class Adservice implements Function {

	private static Logger logger = Logger.getGlobal();
	
	private static int MAX_ADS_TO_SERVE = 2;
	
	private static final ImmutableListMultimap<String, Ad> adsMap = createAdsMap();
	
	private static final Random random = new Random();
	
	public ResponseEntity call(RequestEntity req, Context context) {
		List<String> context_keys = new ArrayList<String>();
		URI url = req.getUrl();
		String query = url.getQuery();
		if (query != null) {
			String[] pairs = query.split("&");
			for (String pair : pairs) {
				int idx = pair.indexOf("=");
				try {
					String key = URLDecoder.decode(pair.substring(0, idx), "UTF-8");
					if (!key.equals("context_keys")) {
						logger.log(Level.SEVERE, "the key is incorrect: "+key);
						return ResponseEntity.status(HttpStatus.BAD_REQUEST).build();
					}
					String value =  URLDecoder.decode(pair.substring(idx + 1), "UTF-8");
					Collections.addAll(context_keys, value.split(","));
				} catch (UnsupportedEncodingException e) {
					logger.log(Level.SEVERE, "failed decode");
					e.printStackTrace();
					return ResponseEntity.status(HttpStatus.BAD_REQUEST).build();
				}
			}
		}
		AdResponse res = getAds(new AdRequest(context_keys));
		HttpHeaders headers = new HttpHeaders();
		headers.add("content-type", "application/json");
		ObjectMapper objectMapper = new ObjectMapper();
		String body = null;
		try {
			body = objectMapper.writeValueAsString(res);
		} catch (JsonProcessingException e) {
			logger.log(Level.SEVERE, "failed serialize AdResponse");
			e.printStackTrace();
		}
		logger.log(Level.INFO, "successfully get adverts");
		return ResponseEntity.status(HttpStatus.OK).headers(headers).body(body);
	}
	
    public AdResponse getAds(AdRequest req) {
        try {
          List<Ad> allAds = new ArrayList<Ad>();
          logger.info("received ad request (context_words=" + req.getContextKeysList() + ")");
          if (req.getContextKeysCount() > 0) {
            for (int i = 0; i < req.getContextKeysCount(); i++) {
              Collection<Ad> ads = getAdsByCategory(req.getContext_keys()[i]);
              allAds.addAll(ads);
            }
          } else {
            logger.info("No Context provided. Constructing random Ads.");
            allAds = getRandomAds();
          }
          if (allAds.isEmpty()) {
            // Serve random ads.
            logger.info("No Ads found based on context. Constructing random Ads.");
            allAds = getRandomAds();
          }
          AdResponse res = new AdResponse(allAds);
          return res;
        } catch (Exception e) {
          e.printStackTrace();
          logger.log(Level.SEVERE, "GetAds Failed");
        }
		return null;
      }

    private Collection<Ad> getAdsByCategory(String category) {
        return adsMap.get(category);
    }
    
    private static ImmutableListMultimap<String, Ad> createAdsMap() {
        Ad hairdryer = new Ad("/product/2ZYFJ3GM2N","Hairdryer for sale. 50% off.");
        Ad tankTop = new Ad("/product/66VCHSJNUP","Tank top for sale. 20% off.");
        Ad candleHolder = new Ad("/product/0PUK6V6EV0", "Candle holder for sale. 30% off.");
        Ad bambooGlassJar = new Ad("/product/9SIQT8TOJO","Bamboo glass jar for sale. 10% off.");
        Ad watch = new Ad("/product/1YMWWN1N4O","Watch for sale. Buy one, get second kit for free");
        Ad mug = new Ad("/product/6E92ZMYYFZ","Mug for sale. Buy two, get third one for free");
        Ad loafers = new Ad("/product/L9ECAV7KIM","Loafers for sale. Buy one, get second one for free");
        return ImmutableListMultimap.<String, Ad>builder()
            .putAll("clothing", tankTop)
            .putAll("accessories", watch)
            .putAll("footwear", loafers)
            .putAll("hair", hairdryer)
            .putAll("decor", candleHolder)
            .putAll("kitchen", bambooGlassJar, mug)
            .build();
    }
    
    private List<Ad> getRandomAds() {
        List<Ad> ads = new ArrayList<Ad>(MAX_ADS_TO_SERVE);
        Collection<Ad> allAds = adsMap.values();
        for (int i = 0; i < MAX_ADS_TO_SERVE; i++) {
          ads.add(Iterables.get(allAds, random.nextInt(allAds.size())));
        }
        return ads;
    }
  
}

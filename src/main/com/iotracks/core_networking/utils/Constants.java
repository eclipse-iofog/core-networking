package main.com.iotracks.core_networking.utils;

import io.netty.util.CharsetUtil;

/**
 * Created by saeid on 4/12/16.
 */
public class Constants {
    public static byte[] BEAT = "BEAT".getBytes(CharsetUtil.US_ASCII);
    public static byte[] ACK = "ACK".getBytes(CharsetUtil.US_ASCII);
    public static byte[] TXEND = "TXEND".getBytes(CharsetUtil.US_ASCII);
}

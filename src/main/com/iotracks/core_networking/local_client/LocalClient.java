package main.com.iotracks.core_networking.local_client;

/**
 * Created by saeid on 4/13/16.
 */
public interface LocalClient {
    void sendMessage(byte[] message);

    boolean connect(long timeout);

    boolean isConnected();
}

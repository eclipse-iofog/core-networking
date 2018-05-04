package main.org.eclipse.iofog.core_networking.local_client;

/**
 * interface for private and public modes
 * <p>
 * Created by saeid on 4/13/16.
 */
public interface LocalClient {
    void handleMessage(byte[] message);

    boolean connect(long timeout);

    boolean isConnected();

    void closeConnection();
}

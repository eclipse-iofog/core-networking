package main.org.eclipse.iofog.core_networking.comsat_client;

import java.util.concurrent.ThreadFactory;

/**
 * Thread factory for ComSat clients
 * <p>
 * Created by saeid on 4/11/16.
 */
public class ComSatClientThreadFactory implements ThreadFactory {
    @Override
    public Thread newThread(Runnable runnable) {
        Thread result = new Thread(runnable, String.format("ComSatClient #%d", ((ComSatClient) runnable).getConnectionId()));
        return result;
    }
}

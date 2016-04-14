package main.com.iotracks.core_networking.utils;

import com.iotracks.elements.IOMessage;

import java.util.LinkedList;
import java.util.Queue;

/**
 * repository for received {@link IOMessage} from local container
 *
 * Created by saeid on 4/13/16.
 */
public class MessageRepository {
    public static Object messageLock = new Object();
    private static Queue<IOMessage> messages = new LinkedList<>();

    /**
     * adds received {@link IOMessage} to messages queue
     *
     * @param message
     */
    public static synchronized void pushMessage(IOMessage message) {
        messages.offer(message);
        synchronized (messageLock) {
            messageLock.notify();
        }
    }

    /**
     * returns {@link IOMessage} on top of the messages queue
     *
     * @return
     */
    public static synchronized IOMessage peekMessage() {
        return messages.peek();
    }

    /**
     * removes {@link IOMessage} from top of the messages queue
     *
     */
    public static synchronized void removeHead() {
        messages.poll();
    }
}

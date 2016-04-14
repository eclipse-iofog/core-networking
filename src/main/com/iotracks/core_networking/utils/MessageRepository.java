package main.com.iotracks.core_networking.utils;

import com.iotracks.elements.IOMessage;

import java.util.LinkedList;
import java.util.Queue;

/**
 * Created by saeid on 4/13/16.
 */
public class MessageRepository {
    private static Queue<IOMessage> messages = new LinkedList<>();
    public static Object messageLock = new Object();

    public static synchronized void pushMessage(IOMessage message) {
        messages.offer(message);
        synchronized (messageLock) {
            messageLock.notify();
        }
    }

    public static synchronized IOMessage peekMessage() {
        return messages.peek();
    }

    public static synchronized void removeHead() {
        messages.poll();
    }
}

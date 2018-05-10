package org.eclipse.iofog.core.networking.utils;

import org.eclipse.iofog.elements.IOMessage;

import java.util.LinkedList;
import java.util.Queue;
import java.util.logging.Logger;

/**
 * repository for received {@link IOMessage} from local container
 * <p>
 * Created by saeid on 4/13/16.
 */
public class MessageRepository {
    private static final Logger log = Logger.getLogger(MessageRepository.class.getName());
    public static Object messageLock = new Object();
    private static Queue<IOMessage> messages = new LinkedList<>();

    /**
     * adds received {@link IOMessage} to messages queue
     *
     * @param message
     */
    public static synchronized void pushMessage(IOMessage message) {
        log.info("iomessage added to repository");
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
        log.info("giving iomessage to client handler");
        return messages.peek();
    }

    /**
     * removes {@link IOMessage} from top of the messages queue
     */
    public static synchronized void removeHead() {
        log.info("removing iomessage from repository");
        messages.poll();
    }
}

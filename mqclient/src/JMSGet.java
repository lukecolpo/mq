import com.ibm.msg.client.jms.JmsConnectionFactory;
import com.ibm.msg.client.jms.JmsFactoryFactory;
import com.ibm.msg.client.wmq.WMQConstants;

import javax.jms.*;

public class JMSGet {

    // System exit status value (assume unset value to be 1)
    private static int status = 1;

    private static final String HOST = "0.0.0.0";
    private static final int PORT = 1414;
    private static final String CHANNEL = "DEV.APP.SVRCONN";
    private static final String QMGR = "QM1";
    private static final String APP_USER = "app";
    private static final String APP_PASSWORD = "_APP_PASSWORD_";
    private static final String QUEUE_NAME = "LUKE.ALIAS.QUEUE";

    public static void main(String[] args) {
        // Variables
        JMSContext context = null;
        Destination destination = null;
        JMSProducer producer = null;
        JMSConsumer consumer = null;

        try {
            // Create a connection factory
            JmsFactoryFactory ff = JmsFactoryFactory.getInstance(WMQConstants.WMQ_PROVIDER);
            JmsConnectionFactory cf = ff.createConnectionFactory();

            // Set the properties
            cf.setStringProperty(WMQConstants.WMQ_HOST_NAME, HOST);
            cf.setIntProperty(WMQConstants.WMQ_PORT, PORT);
            cf.setStringProperty(WMQConstants.WMQ_CHANNEL,CHANNEL);
            cf.setIntProperty(WMQConstants.WMQ_CONNECTION_MODE, WMQConstants.WMQ_CM_CLIENT);
            cf.setStringProperty(WMQConstants.WMQ_QUEUE_MANAGER, QMGR);
            cf.setStringProperty(WMQConstants.WMQ_APPLICATIONNAME, "JmsPutGet (JMS)");
            cf.setBooleanProperty(WMQConstants.USER_AUTHENTICATION_MQCSP,true);
            cf.setStringProperty(WMQConstants.USERID, APP_USER);
            cf.setStringProperty(WMQConstants.PASSWORD, APP_PASSWORD);

            // Create JMS Objects
            context = cf.createContext();
            destination = context.createQueue("queue:///" + QUEUE_NAME);

            long uniqueNumber = System.currentTimeMillis() % 1000;
            TextMessage message = context.createTextMessage("testing 1 2 .."+ uniqueNumber);

            consumer = context.createConsumer(destination);
            String receivedMessage = consumer.receiveBody(String.class, 15000);
            System.out.println("\nReceived Message\n");

            recordSuccess();
        } catch(JMSException jmsex) {
            recordFailure(jmsex);
        }
    }

    private static void recordSuccess() {
        System.out.println("SUCCESS");
        status = 0;
        return;
    }

    private static void recordFailure(Exception ex) {
        if(ex != null) {
            if (ex instanceof JMSException) {
                processJMSException((JMSException) ex);
            } else {
                System.out.println(ex);
            }
        }
        System.out.println("FAILURE");
        status = -1;
    }

    private static void processJMSException(JMSException ex) {
        System.out.println(ex);
        Throwable innerException = ex.getLinkedException();
        if (innerException != null) {
            System.out.println("Inner exception(S):");
        }
        while (innerException != null) {
            System.out.println(innerException);
            innerException = innerException.getCause();
        }
    }
}
